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
	rfv1beta3 "github.com/refunc/refunc/pkg/apis/refunc/v1beta3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func UpdateURL(c *gin.Context) {
	functionName := c.Param("FunctionName")
	triggerName := fmt.Sprintf("lambda-http-%s", functionName)
	var payload apis.FunctionURLConfig
	if err := c.BindJSON(&payload); err != nil {
		awsutils.AWSErrorResponse(c, 400, "InvalidParameterValueException")
		return
	}

	refuncClient, err := utils.GetRefuncClient(c)
	if err != nil {
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	region := c.GetString("region")

	currentTrigger, err := refuncClient.RefuncV1beta3().Triggers(region).Get(context.TODO(), triggerName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("get current trigger error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	if errors.IsNotFound(err) {
		awsutils.AWSErrorResponse(c, 404, "ResourceNotFoundException")
		return
	}

	currentTrigger.Spec.TriggerConfig = rfv1beta3.TriggerConfig{
		HTTP: &rfv1beta3.HTTPTrigger{
			AuthType: "None",
			Cors:     rfv1beta3.HTTPTriggerCors(payload.Cors),
		},
	}

	updatedTrigger, err := refuncClient.RefuncV1beta3().Triggers(region).Update(context.TODO(), currentTrigger, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("ensure trigger error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	urlConfig, err := controllers.HTTPtriggerToURLConfig(*updatedTrigger)
	if err != nil {
		klog.Errorf("httptrigger to lambda url config error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	c.JSON(http.StatusOK, urlConfig)
}
