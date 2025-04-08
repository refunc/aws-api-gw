package eventsourcemapping

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/refunc/aws-api-gw/pkg/apis"
	"github.com/refunc/aws-api-gw/pkg/controllers"
	"github.com/refunc/aws-api-gw/pkg/utils"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func GetEventSource(c *gin.Context) {
	triggerName := c.Param("EventSourceName")

	refuncClient, err := utils.GetRefuncClient(c)
	if err != nil {
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	region := c.GetString("region")
	trigger, err := refuncClient.RefuncV1beta3().Triggers(region).Get(context.TODO(), triggerName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("get trigger error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	if errors.IsNotFound(err) {
		awsutils.AWSErrorResponse(c, 404, "ResourceNotFoundException")
		return
	}

	eventConfig, err := controllers.TriggerToEventSourceConfig(*trigger)
	if err != nil {
		klog.Errorf("trigger to lambda event source error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	c.JSON(http.StatusOK, eventConfig)
}

func ListEventSource(c *gin.Context) {
	arnInfo := strings.Split(c.Query("EventSourceArn"), ":")
	if len(arnInfo) < 2 {
		awsutils.AWSErrorResponse(c, 400, "InvalidParameterValueException")
		return
	}
	triggerType, err := controllers.GetArnTriggerType(arnInfo[1])
	if err != nil {
		klog.Errorf("arn info params error %v", err)
		awsutils.AWSErrorResponse(c, 400, "InvalidParameterValueException")
		return
	}
	if triggerType == controllers.HTTPTriggerType {
		awsutils.AWSErrorResponse(c, 400, "InvalidParameterValueException HTTPTriger Managed by FunctionURL")
		return
	}
	options := metav1.ListOptions{
		LabelSelector: controllers.LambdaLabelTriggerType + "=" + triggerType + "," + controllers.LambdaLabelFuncdef + "=" + c.Query("FunctionName"),
	}
	if triggerType == "*" {
		// list any type trigger, except http trigger
		options = metav1.ListOptions{
			LabelSelector: controllers.LambdaLabelTriggerType + "!=" + controllers.HTTPTriggerType + "," + controllers.LambdaLabelFuncdef + "=" + c.Query("FunctionName"),
		}
	}
	limit, err := strconv.Atoi(c.Query("MaxItems"))
	if err == nil && limit > 0 {
		options.Limit = int64(limit)
	}
	marker := c.Query("Marker")
	if marker != "" {
		options.Continue = marker
	}

	refuncClient, err := utils.GetRefuncClient(c)
	if err != nil {
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	region := c.GetString("region")
	triggers, err := refuncClient.RefuncV1beta3().Triggers(region).List(context.TODO(), options)
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("list triggers error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	if errors.IsNotFound(err) {
		awsutils.AWSErrorResponse(c, 404, "ResourceNotFoundException")
		return
	}

	events := []apis.EventSourceMappingConfiguration{}

	for _, item := range triggers.Items {
		eventConfig, err := controllers.TriggerToEventSourceConfig(item)
		if err != nil {
			klog.Errorf("trigger to lambda configuration error %v", err)
			awsutils.AWSErrorResponse(c, 500, "ServiceException")
			return
		}
		events = append(events, eventConfig)
	}

	c.JSON(http.StatusOK, apis.ListEventSourceMappingResponse{
		EventSourceMappings: events,
		NextMarker:          triggers.ListMeta.Continue,
	})
}
