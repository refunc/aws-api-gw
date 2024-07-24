package eventsourcemapping

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/refunc/aws-api-gw/pkg/utils"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func DeleteEventSource(c *gin.Context) {
	triggerName := c.Param("EventSourceName")

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

	err = refuncClient.RefuncV1beta3().Triggers(region).Delete(context.TODO(), trigger.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("delete trigger error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	c.AbortWithStatus(204)
}
