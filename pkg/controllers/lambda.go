package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refunc/aws-api-gw/pkg/apis"
	"github.com/refunc/aws-api-gw/pkg/utils"
)

func CreateFunction(c *gin.Context) {
	var payload apis.CreateFunctionRequest
	var rsp apis.CreateFunctionResponse
	if err := c.BindJSON(&payload); err != nil {
		utils.AWSErrorResponse(c, 400, "InvalidParameterValueException")
		return
	}
	// kubeClient, err := utils.GetKubeClient(c)
	// if err != nil {
	// 	klog.Error(err)
	// 	utils.AWSErrorResponse(c, 500, "ServiceException")
	// 	return
	// }
	utils.LogObject(payload)
	c.JSON(http.StatusOK, rsp)
}
