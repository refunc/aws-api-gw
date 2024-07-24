package urls

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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

func CreateURL(c *gin.Context) {
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

	funcdef, err := refuncClient.RefuncV1beta3().Funcdeves(region).Get(context.TODO(), functionName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("get funcdef error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	if errors.IsNotFound(err) {
		awsutils.AWSErrorResponse(c, 404, "ResourceNotFoundException")
		return
	}

	trigger, err := refuncClient.RefuncV1beta3().Triggers(region).Create(context.TODO(), &rfv1beta3.Trigger{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: funcdef.Namespace,
			Name:      triggerName,
			Labels: map[string]string{
				controllers.LambdaLabelAutoCreated: "true",
				controllers.LambdaLabelFuncdef:     funcdef.Name,
				controllers.LambdaLabelTriggerType: controllers.HTTPTriggerType,
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
			Type:     controllers.HTTPTriggerType,
			FuncName: funcdef.Name,
			TriggerConfig: rfv1beta3.TriggerConfig{
				HTTP: &rfv1beta3.HTTPTrigger{
					AuthType: "None",
					Cors:     rfv1beta3.HTTPTriggerCors(payload.Cors),
				},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("create trigger error %v", err)
		if strings.Contains(err.Error(), "exists") {
			awsutils.AWSErrorResponse(c, 409, "ResourceConflictException")
		} else {
			awsutils.AWSErrorResponse(c, 500, "ServiceException")
		}
		return
	}

	urlConfig, err := controllers.HTTPtriggerToURLConfig(*trigger)
	if err != nil {
		klog.Errorf("httptrigger to lambda url config error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	c.JSON(http.StatusOK, urlConfig)
}
