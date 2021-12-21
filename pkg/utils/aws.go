package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AWSErrorResponse(c *gin.Context, code int, msg string) {
	theType := "User"
	if code >= 500 {
		theType = "Server"
	}
	errorType := "InternalFailure"
	c.Header("x-amzn-errortype", errorType)
	c.JSON(http.StatusBadRequest, gin.H{
		"Type":    theType,
		"message": msg,
		"__type":  errorType,
	})
}
