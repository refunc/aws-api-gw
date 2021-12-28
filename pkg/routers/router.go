package routers

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/refunc/aws-api-gw/pkg/controllers"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	awsCredentials "github.com/aws/aws-sdk-go/aws/credentials"
	awsSigner "github.com/aws/aws-sdk-go/aws/signer/v4"

	"github.com/gin-gonic/gin"
	"github.com/refunc/refunc/pkg/env"
	"github.com/refunc/refunc/pkg/utils/cmdutil/sharedcfg"
)

func CreateHTTPRouter(sc sharedcfg.Configs, cfg Config, stopC <-chan struct{}) *gin.Engine {

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(WithClientSet(sc, stopC))
	router.Use(WithAwsSign(sc, cfg.Rbac))
	apis := router.Group("/2015-03-31")
	{
		apis.POST("/functions", controllers.CreateFunction)
		apis.GET("/functions/", controllers.ListFunction)
		apis.GET("/functions/:FunctionName", controllers.GetFunction)
		apis.DELETE("/functions/:FunctionName", controllers.DeleteFunction)
		apis.PUT("/functions/:FunctionName/code", controllers.UpdateFunctionCode)
		apis.PUT("/functions/:FunctionName/configuration", controllers.UpdateFunctionConfiguration)
		apis.POST("/functions/:FunctionName/invocations", controllers.InvokeFunction)
	}
	return router
}

func WithClientSet(sc sharedcfg.Configs, stopC <-chan struct{}) gin.HandlerFunc {
	kubeClient := sc.KubeClient()
	refuncClient := sc.RefuncClient()
	kubeInformers := sc.KubeInformers()
	refuncInformers := sc.RefuncInformers()
	refuncFundefLister := refuncInformers.Refunc().V1beta3().Funcdeves().Lister()
	serviceAccountLister := kubeInformers.Core().V1().ServiceAccounts().Lister()
	wantedInformers := []cache.InformerSynced{
		refuncInformers.Refunc().V1beta3().Funcdeves().Informer().HasSynced,
		kubeInformers.Core().V1().ServiceAccounts().Informer().HasSynced,
		kubeInformers.Core().V1().Secrets().Informer().HasSynced,
	}

	go func() {
		if !cache.WaitForCacheSync(stopC, wantedInformers...) {
			klog.Fatalln("Fail wait for cache sync")
		}
		klog.Infoln("Success sync informer cache")
	}()

	hostname, err := os.Hostname()
	if err != nil {
		klog.Fatalf("get hostname error %v", err)
	}
	natsConn, err := env.NewNatsConn(nats.Name(os.Getenv("aws-api-gw") + "/" + hostname))
	if err != nil {
		klog.Fatalf("connect to nats error %v", err)
	}
	return func(c *gin.Context) {
		c.Set("kc", kubeClient)
		c.Set("rc", refuncClient)
		c.Set("funcdefLister", refuncFundefLister)
		c.Set("serviceAccountLister", serviceAccountLister)
		c.Set("nats", natsConn)
		c.Next()
	}
}

func WithAwsSign(sc sharedcfg.Configs, rbac bool) gin.HandlerFunc {

	kubeInformers := sc.KubeInformers()
	serviceAccountLister := kubeInformers.Core().V1().ServiceAccounts().Lister()
	secretLister := kubeInformers.Core().V1().Secrets().Lister()

	ns := sc.Namespace()
	reg := regexp.MustCompile(`Credential=(.*\/.*\/.*\/lambda/aws4_request), SignedHeaders`)
	return func(c *gin.Context) {
		amzDate := c.Request.Header.Get("X-Amz-Date")
		dt, err := time.Parse(timeFormat, amzDate)
		if err != nil {
			klog.Error(err)
			awsutils.AWSErrorResponse(c, 400, "InvalidCredentialException")
			c.Abort()
			return
		}

		authorization := c.Request.Header.Get(authorizationHeader)
		if authorization == "" {
			awsutils.AWSErrorResponse(c, 400, "InvalidAuthorizationException")
			c.Abort()
			return
		}

		matches := reg.FindStringSubmatch(authorization)
		if len(matches) != 2 {
			awsutils.AWSErrorResponse(c, 400, "InvalidCredentialException")
			c.Abort()
			return
		}
		credentials := strings.Split(matches[1], "/")
		if len(credentials) != 5 {
			awsutils.AWSErrorResponse(c, 400, "InvalidCredentialException")
			c.Abort()
			return
		}
		region := credentials[2]
		if ns != "" && ns != region {
			awsutils.AWSErrorResponse(c, 400, "InvalidRegionException")
			c.Abort()
			return
		}
		if region == "" {
			awsutils.AWSErrorResponse(c, 400, "InvalidRegionException")
			c.Abort()
			return
		}

		accessKeyID := credentials[0]
		if rbac {
			// gen access_key_id and access_secret base on serviceaccount
			sa, err := serviceAccountLister.ServiceAccounts(region).Get(accessKeyID)
			if err != nil {
				klog.Error(err)
				awsutils.AWSErrorResponse(c, 400, "InvalidCredentialException")
				c.Abort()
				return
			}
			if len(sa.Secrets) != 1 {
				klog.Errorf("secret for serviceaccount %s/%s error", sa.Namespace, sa.Name)
				awsutils.AWSErrorResponse(c, 400, "InvalidCredentialException")
				c.Abort()
				return
			}
			secret, err := secretLister.Secrets(region).Get(sa.Secrets[0].Name)
			if err != nil {
				klog.Error(err)
				awsutils.AWSErrorResponse(c, 400, "InvalidCredentialException")
				c.Abort()
				return
			}
			tokenBts, ok := secret.Data["token"]
			if !ok {
				klog.Error(err)
				awsutils.AWSErrorResponse(c, 400, "InvalidCredentialException")
				c.Abort()
				return
			}

			//verify aws signature without body sha256
			accessSecret := base64.StdEncoding.EncodeToString(tokenBts)

			signer := awsSigner.NewSigner(awsCredentials.NewStaticCredentials(accessKeyID, accessSecret, ""))
			signReq, _ := http.NewRequest(c.Request.Method, c.Request.URL.String(), ioutil.NopCloser(nil))
			signReq.URL, signReq.Host = c.Request.URL, c.Request.Host
			_, err = signer.Sign(signReq, nil, "lambda", region, dt)
			if err != nil {
				klog.Error(err)
				awsutils.AWSErrorResponse(c, 400, "InvalidCredentialException")
				c.Abort()
				return
			}

			if authorization != signReq.Header.Get(authorizationHeader) {
				klog.Infof("verify sign diff (%s) -> (%s)", authorization, signReq.Header.Get(authorizationHeader))
				awsutils.AWSErrorResponse(c, 400, "InvalidCredentialException")
				c.Abort()
				return
			}
		}

		c.Set("region", region)
		c.Next()
	}
}
