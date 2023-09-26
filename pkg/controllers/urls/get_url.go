package urls

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refunc/aws-api-gw/pkg/apis"
	"github.com/refunc/aws-api-gw/pkg/controllers"
	"github.com/refunc/aws-api-gw/pkg/utils"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func GetURL(c *gin.Context) {
	functionName := c.Param("FunctionName")
	triggerName := fmt.Sprintf("lambda-http-%s", functionName)

	refuncClient, err := utils.GetRefuncClient(c)
	if err != nil {
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	region := c.GetString("region")
	trigger, err := refuncClient.RefuncV1beta3().Triggers(region).Get(context.TODO(), triggerName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("get httptrigger error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	if errors.IsNotFound(err) {
		awsutils.AWSErrorResponse(c, 404, "ResourceNotFoundException")
		return
	}

	urlConfig, err := controllers.HTTPriggerToURLConfig(*trigger)
	if err != nil {
		klog.Errorf("httptrigger to lambda url config error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	c.JSON(http.StatusOK, urlConfig)
}

func ListURL(c *gin.Context) {
	functionName := c.Param("FunctionName")
	triggerName := fmt.Sprintf("lambda-http-%s", functionName)

	refuncClient, err := utils.GetRefuncClient(c)
	if err != nil {
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	region := c.GetString("region")
	trigger, err := refuncClient.RefuncV1beta3().Triggers(region).Get(context.TODO(), triggerName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("get httptrigger error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	if errors.IsNotFound(err) {
		awsutils.AWSErrorResponse(c, 404, "ResourceNotFoundException")
		return
	}

	urlConfig, err := controllers.HTTPriggerToURLConfig(*trigger)
	if err != nil {
		klog.Errorf("httptrigger to lambda url config error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	c.JSON(http.StatusOK, apis.ListURLResponse{
		FunctionUrlConfigs: []apis.FunctionURLConfig{urlConfig},
		NextMarker:         "",
	})
}
