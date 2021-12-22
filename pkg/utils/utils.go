package utils

import (
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"
	rfclientset "github.com/refunc/refunc/pkg/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func LogObject(o interface{}) {
	bts, _ := json.Marshal(o)
	klog.Infoln(string(bts))
}

func GetKubeClient(c *gin.Context) (kubernetes.Interface, error) {
	kc, ok := c.Get("kc")
	if !ok {
		return nil, errors.New("get kubeClient error")
	}
	return kc.(kubernetes.Interface), nil
}

func GetRefuncClient(c *gin.Context) (rfclientset.Interface, error) {
	kc, ok := c.Get("rc")
	if !ok {
		return nil, errors.New("get refuncClient error")
	}
	return kc.(rfclientset.Interface), nil
}
