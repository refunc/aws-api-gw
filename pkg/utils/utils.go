package utils

import (
	"encoding/json"

	"k8s.io/klog/v2"
)

func LogObject(o interface{}) {
	bts, _ := json.Marshal(o)
	klog.Infoln(string(bts))
}
