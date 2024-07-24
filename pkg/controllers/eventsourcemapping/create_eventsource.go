package eventsourcemapping

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/refunc/aws-api-gw/pkg/apis"
	"github.com/refunc/aws-api-gw/pkg/controllers"
	"github.com/refunc/aws-api-gw/pkg/utils"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
	rfv1beta3 "github.com/refunc/refunc/pkg/apis/refunc/v1beta3"
	"github.com/robfig/cron"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func CreateEventSource(c *gin.Context) {
	var payload apis.EventSourceMappingConfiguration
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

	funcdef, err := refuncClient.RefuncV1beta3().Funcdeves(region).Get(context.TODO(), payload.FunctionArn, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		klog.Errorf("get funcdef error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	if k8serrors.IsNotFound(err) {
		awsutils.AWSErrorResponse(c, 404, "ResourceNotFoundException")
		return
	}

	triggerInfo, err := triggerCutter(funcdef, payload)
	if err != nil {
		klog.Errorf("create trigger info error %v", err)
		awsutils.AWSErrorResponse(c, 400, "InvalidParameterValueException")
		return
	}

	trigger, err := refuncClient.RefuncV1beta3().Triggers(region).Create(context.TODO(), triggerInfo, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("create trigger error %v", err)
		if strings.Contains(err.Error(), "exists") {
			awsutils.AWSErrorResponse(c, 409, "ResourceConflictException")
		} else {
			awsutils.AWSErrorResponse(c, 500, "ServiceException")
		}
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

func triggerCutter(funcdef *rfv1beta3.Funcdef, ec apis.EventSourceMappingConfiguration) (*rfv1beta3.Trigger, error) {
	arns := strings.Split(ec.EventSourceArn, ":")
	if len(arns) != 3 {
		return nil, fmt.Errorf("event arn %s format error", ec.EventSourceArn)
	}
	triggerType, triggerName := arns[1], arns[2]
	//only support cron event.
	if triggerType != "cron" {
		return nil, errors.New("not support event type")
	}
	triggerName = fmt.Sprintf("lambda-cron-%s-%s", ec.FunctionArn, triggerName)
	triggerConfig := rfv1beta3.CronTrigger{}
	for key, vals := range ec.SelfManagedEventSource.Endpoints {
		if key == "cron" && len(vals) > 0 {
			if _, err := cron.Parse(vals[0]); err != nil {
				return nil, err
			}
			triggerConfig.Cron = vals[0]
		}
		if key == "location" && len(vals) > 0 {
			if _, err := time.LoadLocation(vals[0]); err != nil {
				return nil, err
			}
			triggerConfig.Location = vals[0]
		}
		if key == "args" && len(vals) > 0 {
			val := json.RawMessage(vals[0])
			if err := json.Unmarshal(json.RawMessage(vals[0]), &map[string]interface{}{}); err != nil {
				return nil, err
			}
			triggerConfig.Args = val
		}
		if key == "saveLog" && len(vals) > 0 {
			val, err := strconv.ParseBool(vals[0])
			if err != nil {
				return nil, err
			}
			triggerConfig.SaveLog = val
		}
		if key == "saveResult" && len(vals) > 0 {
			val, err := strconv.ParseBool(vals[0])
			if err != nil {
				return nil, err
			}
			triggerConfig.SaveResult = val
		}
	}
	trigger := &rfv1beta3.Trigger{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: funcdef.Namespace,
			Name:      triggerName,
			Labels: map[string]string{
				controllers.LambdaLabelAutoCreated: "true",
				controllers.LambdaLabelFuncdef:     funcdef.Name,
				controllers.LambdaLabelTriggerType: controllers.CronTriggerType,
			},
			Annotations: map[string]string{
				rfv1beta3.AnnotationRPCVer: "v2",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: rfv1beta3.APIVersion,
					Kind:       rfv1beta3.FuncdefKind,
					Name:       funcdef.Name,
					UID:        funcdef.UID,
				},
			},
		},
		Spec: rfv1beta3.TriggerSpec{
			Type:     controllers.CronTriggerType,
			FuncName: funcdef.Name,
			TriggerConfig: rfv1beta3.TriggerConfig{
				Cron: &triggerConfig,
			},
		},
	}
	return trigger, nil
}
