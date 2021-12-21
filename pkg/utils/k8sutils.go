package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/refunc/refunc/pkg/generated/informers/externalversions/refunc"
	"k8s.io/client-go/kubernetes"
)

func GetKubeClient(c *gin.Context) (kubernetes.Interface, error) {
	kc, ok := c.Get("kc")
	if !ok {
		return nil, errors.New("get kubeClient error")
	}
	return kc.(kubernetes.Interface), nil
}

func GetRefuncClient(c *gin.Context) (refunc.Interface, error) {
	kc, ok := c.Get("rc")
	if !ok {
		return nil, errors.New("get refuncClient error")
	}
	return kc.(refunc.Interface), nil
}
