package routers

import (
	"regexp"
	"strings"

	"github.com/refunc/aws-api-gw/pkg/controllers"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"

	"github.com/gin-gonic/gin"
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
		apis.GET("/functions/:FunctionName", controllers.GetFunction)
	}
	return router
}

func WithClientSet(sc sharedcfg.Configs) gin.HandlerFunc {
	kubeClient := sc.KubeClient()
	refuncClient := sc.RefuncClient()
	return func(c *gin.Context) {
		c.Set("kc", kubeClient)
		c.Set("rc", refuncClient)
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
			return
		}
		//TODO Verify Authorization Signature
		matches := reg.FindStringSubmatch(authorization)
		if len(matches) != 2 {
			awsutils.AWSErrorResponse(c, 400, "InvalidCredentialException")
			return
		}
		credentials := strings.Split(matches[1], "/")
		if len(credentials) != 5 {
			awsutils.AWSErrorResponse(c, 400, "InvalidCredentialException")
			return
		}
		region := credentials[2]
		if ns != "" && ns != region {
			awsutils.AWSErrorResponse(c, 400, "InvalidRegionException")
			return
		}
		if region == "" {
			awsutils.AWSErrorResponse(c, 400, "InvalidRegionException")
			return
		}
		c.Set("region", region)
		c.Next()
	}
}
