package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/refunc/aws-api-gw/pkg/apis"
	"github.com/refunc/aws-api-gw/pkg/controllers"
	"github.com/refunc/aws-api-gw/pkg/services"
	"github.com/refunc/aws-api-gw/pkg/utils"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
	"github.com/refunc/aws-api-gw/pkg/utils/rfutils"
	rfv1beta3 "github.com/refunc/refunc/pkg/apis/refunc/v1beta3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog/v2"
)

func CreateFunction(c *gin.Context) {
	var payload apis.CreateFunctionRequest
	if err := c.BindJSON(&payload); err != nil {
		awsutils.AWSErrorResponse(c, 400, "InvalidParameterValueException")
		return
	}

	if errs := validation.IsDNS1123Label(payload.FunctionName); len(errs) > 0 {
		awsutils.AWSErrorResponse(c, 400, "InvalidFunctionNameException")
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
			Labels: map[string]string{
				rfv1beta3.LabelLambdaName:    payload.FunctionName,
				rfv1beta3.LabelLambdaVersion: controllers.LambdaVersion,
			},
			Annotations: map[string]string{},
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

	fnConfiguration, err := controllers.FuncdefToLambdaConfiguration(*funcdef)
	if err != nil {
		klog.Errorf("funcdef to lambda configuration error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	c.JSON(http.StatusOK, apis.CreateFunctionResponse{
		FunctionConfiguration: fnConfiguration,
	})
}
