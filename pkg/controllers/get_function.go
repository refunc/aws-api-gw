package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refunc/aws-api-gw/pkg/apis"
	"github.com/refunc/aws-api-gw/pkg/utils"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func GetFunction(c *gin.Context) {
	functionName := c.Param("FunctionName")
	//TODO support get function Qualifier
	refuncClient, err := utils.GetRefuncClient(c)
	if err != nil {
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	region := c.GetString("region")
	fndef, err := refuncClient.RefuncV1beta3().Funcdeves(region).Get(context.TODO(), functionName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	if errors.IsNotFound(err) {
		awsutils.AWSErrorResponse(c, 404, "ResourceNotFoundException")
		return
	}

	fnConfiguration, err := FuncdefToLambdaConfiguration(*fndef)
	if err != nil {
		klog.Errorf("funcdef to lambda configuration error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	c.JSON(http.StatusOK, apis.GetFunctionResponse{
		Code: map[string]string{
			"Location": fndef.Spec.Body,
		},
		Configuration: fnConfiguration,
	})
}
