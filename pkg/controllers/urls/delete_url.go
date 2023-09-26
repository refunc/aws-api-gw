package urls

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/refunc/aws-api-gw/pkg/utils"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func DeleteURL(c *gin.Context) {
	functionName := c.Param("FunctionName")
	triggerName := fmt.Sprintf("lambda-http-%s", functionName)
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

	err = refuncClient.RefuncV1beta3().Triggers(region).Delete(context.TODO(), currentTrigger.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("delete trigger error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	c.AbortWithStatus(204)
}
