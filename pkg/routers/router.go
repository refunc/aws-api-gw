package routers

import (
	"os"
	"regexp"
	"strings"

	nats "github.com/nats-io/nats.go"
	"github.com/refunc/aws-api-gw/pkg/controllers"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
	"k8s.io/klog/v2"

	"github.com/gin-gonic/gin"
	"github.com/refunc/refunc/pkg/env"
	"github.com/refunc/refunc/pkg/utils/cmdutil/sharedcfg"
)

func CreateHTTPRouter(sc sharedcfg.Configs) *gin.Engine {

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(WithAwsSign(sc))
	router.Use(WithClientSet(sc))
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

func WithClientSet(sc sharedcfg.Configs) gin.HandlerFunc {
	kubeClient := sc.KubeClient()
	refuncClient := sc.RefuncClient()
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
		c.Set("nats", natsConn)
		c.Next()
	}
}

func WithAwsSign(sc sharedcfg.Configs) gin.HandlerFunc {
	ns := sc.Namespace()
	reg := regexp.MustCompile(`Credential=(.*\/.*\/.*\/lambda/aws4_request), SignedHeaders`)
	return func(c *gin.Context) {
		authorization := c.Request.Header.Get("Authorization")
		if authorization == "" {
			awsutils.AWSErrorResponse(c, 400, "InvalidAuthorizationException")
			c.Abort()
			return
		}
		//TODO Verify Authorization Signature
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
		c.Set("region", region)
		c.Next()
	}
}
