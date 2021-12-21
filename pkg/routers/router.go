package routers

import (
	"github.com/refunc/aws-api-gw/pkg/controllers"

	"github.com/gin-gonic/gin"
	"github.com/refunc/refunc/pkg/utils/cmdutil/sharedcfg"
)

func CreateHTTPRouter(sc sharedcfg.Configs) *gin.Engine {

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(WithKubeClient(sc))
	apis := router.Group("/2015-03-31")
	{
		apis.POST("/functions", controllers.CreateFunction)
	}
	return router
}

func WithKubeClient(sc sharedcfg.Configs) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("kc", sc.KubeClient())
		c.Set("rc", sc.RefuncClient())
		c.Next()
	}
}
