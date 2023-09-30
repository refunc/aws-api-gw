package functions

import (
	"context"
	"net/http"
	"strconv"

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

func GetFunction(c *gin.Context) {
	functionName := c.Param("FunctionName")
	//TODO support get function Qualifier

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

	fnConfiguration, err := controllers.FuncdefToLambdaConfiguration(*fndef)
	if err != nil {
		klog.Errorf("funcdef to lambda configuration error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	concurrencyNum := 1
	concurrencySpec, ok := fndef.Annotations[rfv1beta3.AnnotationLambdaConcurrency]
	if ok {
		if num, err := strconv.Atoi(concurrencySpec); err == nil {
			concurrencyNum = num
		}
	}

	c.JSON(http.StatusOK, apis.GetFunctionResponse{
		Code: map[string]string{
			"Location": fndef.Spec.Body,
		},
		Configuration: fnConfiguration,
		Concurrency: apis.FunctionConcurrencyConfig{
			ReservedConcurrentExecutions: int64(concurrencyNum),
		},
	})
}

func ListFunction(c *gin.Context) {
	//TODO support list function FunctionVersion MasterRegion
	options := metav1.ListOptions{}
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
	fndeves, err := refuncClient.RefuncV1beta3().Funcdeves(region).List(context.TODO(), options)
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("list funcdef error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	if errors.IsNotFound(err) {
		awsutils.AWSErrorResponse(c, 404, "ResourceNotFoundException")
		return
	}

	functions := []apis.FunctionConfiguration{}

	for _, fndef := range fndeves.Items {
		fnConfiguration, err := controllers.FuncdefToLambdaConfiguration(fndef)
		if err != nil {
			klog.Errorf("funcdef to lambda configuration error %v", err)
			awsutils.AWSErrorResponse(c, 500, "ServiceException")
			return
		}
		functions = append(functions, fnConfiguration)
	}

	c.JSON(http.StatusOK, apis.ListFunctionResponse{
		Functions:  functions,
		NextMarker: fndeves.ListMeta.Continue,
	})
}
