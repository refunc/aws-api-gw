package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/refunc/aws-api-gw/pkg/apis"
	"github.com/refunc/aws-api-gw/pkg/services"
	"github.com/refunc/aws-api-gw/pkg/utils"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
	"github.com/refunc/aws-api-gw/pkg/utils/rfutils"
	rfv1beta3 "github.com/refunc/refunc/pkg/apis/refunc/v1beta3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func CreateFunction(c *gin.Context) {
	var payload apis.CreateFunctionRequest
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
	body, codeSize, hash, err := services.SetFunctionCode(payload.Code, region, payload.FunctionName)
	if err != nil {
		klog.Errorf("set function code error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	timeout := rfutils.GetTimeout(int(payload.Timeout))

	// apply funcdef
	funcdef, err := refuncClient.RefuncV1beta3().Funcdeves(region).Create(context.TODO(), &rfv1beta3.Funcdef{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rfv1beta3.APIVersion,
			Kind:       rfv1beta3.FuncdefKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      payload.FunctionName,
			Namespace: region,
			Annotations: map[string]string{
				LambdaLabelName:    payload.FunctionName,
				LambdaLabelVersion: LambdaVersion,
			},
		},
		Spec: rfv1beta3.FuncdefSpec{
			Body:  body,
			Hash:  hash,
			Entry: payload.Handler,
			Runtime: &rfv1beta3.Runtime{
				Name:    payload.Runtime,
				Envs:    payload.Environment.Variables,
				Timeout: timeout,
			},
			Custom: json.RawMessage(fmt.Sprintf(`{"codeSize":%d}`, codeSize)),
		},
	}, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("create funcdef error %v", err)
		if strings.Contains(err.Error(), "exists") {
			awsutils.AWSErrorResponse(c, 409, "ResourceConflictException")
		} else {
			awsutils.AWSErrorResponse(c, 500, "ServiceException")
		}
		return
	}

	// auto create http trigger
	triggerName := fmt.Sprintf("lambda-http-%s", funcdef.Name)
	trigger := &rfv1beta3.Trigger{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: funcdef.Namespace,
			Name:      triggerName,
			Labels: map[string]string{
				LambdaLabelAutoCreated: "true",
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
			Type:     TriggerType,
			FuncName: funcdef.Name,
		},
	}

	// apply trigger
	currentTrigger, err := refuncClient.RefuncV1beta3().Triggers(region).Get(context.TODO(), triggerName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("get current trigger error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	if currentTrigger.Name == "" {
		_, err = refuncClient.RefuncV1beta3().Triggers(region).Create(context.TODO(), trigger, metav1.CreateOptions{})
	} else {
		trigger.ResourceVersion = currentTrigger.ResourceVersion
		_, err = refuncClient.RefuncV1beta3().Triggers(region).Update(context.TODO(), trigger, metav1.UpdateOptions{})
	}
	if err != nil {
		klog.Errorf("ensure trigger error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	fnConfiguration, err := FuncdefToLambdaConfiguration(*funcdef)
	if err != nil {
		klog.Errorf("funcdef to lambda configuration error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	c.JSON(http.StatusOK, apis.CreateFunctionResponse{
		FunctionConfiguration: fnConfiguration,
	})
}
