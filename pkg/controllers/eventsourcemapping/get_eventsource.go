package eventsourcemapping

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetEventSource(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}

func ListEventSource(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}
