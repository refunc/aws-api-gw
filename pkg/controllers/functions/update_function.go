package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refunc/aws-api-gw/pkg/apis"
	"github.com/refunc/aws-api-gw/pkg/controllers"
	"github.com/refunc/aws-api-gw/pkg/services"
	"github.com/refunc/aws-api-gw/pkg/utils"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func UpdateFunctionCode(c *gin.Context) {
	functionName := c.Param("FunctionName")
	var payload apis.UpdateFunctionCodeRequest
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
	fndef, err := refuncClient.RefuncV1beta3().Funcdeves(region).Get(context.TODO(), functionName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("get funcdef error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	if errors.IsNotFound(err) {
		awsutils.AWSErrorResponse(c, 404, "ResourceNotFoundException")
		return
	}

	//set function code
	code := map[string]string{}
	if payload.S3Bucket != "" && payload.S3Key != "" {
		code["S3Bucket"] = payload.S3Bucket
		code["S3Key"] = payload.S3Key
	}
	if payload.ZipFile != "" {
		code["ZipFile"] = payload.ZipFile
	}
	body, codeSize, hash, err := services.SetFunctionCode(code, region, functionName)
	if err != nil {
		klog.Errorf("set function code error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	if body != fndef.Spec.Body {
		originBody := fndef.Spec.Body
		go func() {
			if err := services.DelFunctionCode(originBody); err != nil {
				klog.Errorf("del function code error %v", err)
			}
		}()
	}
	fndef.Spec.Body = body
	fndef.Spec.Hash = hash
	fndef.Spec.Custom = json.RawMessage(fmt.Sprintf(`{"codeSize":%d}`, codeSize))

	// apply funcdef
	fndef, err = refuncClient.RefuncV1beta3().Funcdeves(region).Update(context.TODO(), fndef, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("update funcdef code error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	fnConfiguration, err := controllers.FuncdefToLambdaConfiguration(*fndef)
	if err != nil {
		klog.Errorf("funcdef to lambda configuration error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	c.JSON(http.StatusOK, apis.UpdateFunctionCodeResponse{
		FunctionConfiguration: fnConfiguration,
	})
}

func UpdateFunctionConfiguration(c *gin.Context) {
	functionName := c.Param("FunctionName")
	var payload apis.UpdateFunctionConfigurationRequest
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
	fndef, err := refuncClient.RefuncV1beta3().Funcdeves(region).Get(context.TODO(), functionName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("get funcdef error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	if errors.IsNotFound(err) {
		awsutils.AWSErrorResponse(c, 404, "ResourceNotFoundException")
		return
	}

	// update function configuration
	if payload.Handler != "" {
		fndef.Spec.Entry = payload.Handler
	}
	if payload.Runtime != "" {
		fndef.Spec.Runtime.Name = payload.Runtime
	}
	if payload.Timeout > 0 {
		fndef.Spec.Runtime.Timeout = int(payload.Timeout)
	}
	if payload.Environment.Variables != nil && len(payload.Environment.Variables) > 0 {
		fndef.Spec.Runtime.Envs = payload.Environment.Variables
	}

	// apply funcdef
	fndef, err = refuncClient.RefuncV1beta3().Funcdeves(region).Update(context.TODO(), fndef, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("update funcdef configuration error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	fnConfiguration, err := controllers.FuncdefToLambdaConfiguration(*fndef)
	if err != nil {
		klog.Errorf("funcdef to lambda configuration error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	c.JSON(http.StatusOK, apis.UpdateFunctionCodeResponse{
		FunctionConfiguration: fnConfiguration,
	})
}
