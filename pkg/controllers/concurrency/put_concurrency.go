package concurrency

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refunc/aws-api-gw/pkg/apis"
	"github.com/refunc/aws-api-gw/pkg/utils"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
	rfv1beta3 "github.com/refunc/refunc/pkg/apis/refunc/v1beta3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func UpdateFunctionConcurrency(c *gin.Context) {
	functionName := c.Param("FunctionName")
	var payload apis.FunctionConcurrencyConfig
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

	if fndef.Annotations == nil {
		fndef.Annotations = map[string]string{
			rfv1beta3.AnnotationLambdaConcurrency: fmt.Sprintf("%d", payload.ReservedConcurrentExecutions),
		}
	} else {
		fndef.Annotations[rfv1beta3.AnnotationLambdaConcurrency] = fmt.Sprintf("%d", payload.ReservedConcurrentExecutions)
	}
	// apply funcdef
	if _, err := refuncClient.RefuncV1beta3().Funcdeves(region).Update(context.TODO(), fndef, metav1.UpdateOptions{}); err != nil {
		klog.Errorf("update funcdef code error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	c.JSON(http.StatusOK, payload)
}
