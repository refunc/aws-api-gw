package eventsourcemapping

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func DeleteEventSource(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}
