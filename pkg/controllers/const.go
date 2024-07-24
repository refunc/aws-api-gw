package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/refunc/aws-api-gw/pkg/apis"
	rfv1beta3 "github.com/refunc/refunc/pkg/apis/refunc/v1beta3"
)

const (
	LambdaVersion            = "0" //TODO support lambda version apis
	LambdaLabelAutoCreated   = "lambda.refunc.io/auto-created"
	LambdaLabelFuncdef       = "lambda.refunc.io/funcdef"
	LambdaLabelTriggerType   = "lambda.refunc.io/triger-type"
	HTTPTriggerType          = "httptrigger"
	CronTriggerType          = "crontrigger"
	HeaderAmzInvocationType  = "X-Amz-Invocation-Type"
	HeaderAmzLogType         = "X-Amz-Log-Type"
	HeaderAmzClientContext   = "X-Amz-Client-Context"
	HeaderAmzFunctionError   = "X-Amz-Function-Error"
	HeaderAmzLogResult       = "X-Amz-Log-Result"
	HeaderAmzExecutedVersion = "X-Amz-Executed-Version"
)

func FuncdefToLambdaConfiguration(fndef rfv1beta3.Funcdef) (apis.FunctionConfiguration, error) {
	custom := map[string]interface{}{}
	err := json.Unmarshal(fndef.Spec.Custom, &custom)
	var codeSize int64
	if err == nil {
		size, ok := custom["codeSize"].(float64)
		if ok {
			codeSize = int64(size)
		}
	}
	return apis.FunctionConfiguration{
		CodeSha256: fndef.Spec.Hash,
		CodeSize:   codeSize,
		Environment: &apis.FunctionEnvironment{
			Variables: fndef.Spec.Runtime.Envs,
		},
		FunctionName: fndef.Name,
		Handler:      fndef.Spec.Entry,
		LastModified: fndef.CreationTimestamp.Format(time.RFC3339),
		Version:      LambdaVersion,
		RevisionId:   fndef.ResourceVersion,
		Runtime:      fndef.Spec.Runtime.Name,
		Timeout:      int64(fndef.Spec.Runtime.Timeout),
	}, nil
}

func HTTPtriggerToURLConfig(trigger rfv1beta3.Trigger) (apis.FunctionURLConfig, error) {
	if trigger.Spec.Type != "httptrigger" {
		return apis.FunctionURLConfig{}, fmt.Errorf("trigger %s not is http type", trigger.Name)
	}
	httpCfg := rfv1beta3.HTTPTrigger{}
	if trigger.Spec.HTTP != nil {
		httpCfg = *trigger.Spec.HTTP
	}
	return apis.FunctionURLConfig{
		AuthType:         httpCfg.AuthType,
		Cors:             apis.URLCors(httpCfg.Cors),
		FunctionArn:      trigger.Spec.FuncName,
		FunctionUrl:      fmt.Sprintf("/%s/%s", trigger.Namespace, trigger.Spec.FuncName),
		CreationTime:     trigger.CreationTimestamp.Format(time.RFC3339),
		LastModifiedTime: trigger.CreationTimestamp.Format(time.RFC3339),
	}, nil
}

func TriggerToEventSourceConfig(trigger rfv1beta3.Trigger) (apis.EventSourceMappingConfiguration, error) {
	if trigger.Spec.Cron == nil {
		return apis.EventSourceMappingConfiguration{}, fmt.Errorf("trigger %s not is cron type", trigger.Name)
	}
	endpoints := map[string][]string{
		"cron": {trigger.Spec.Cron.Cron},
	}
	if len(trigger.Spec.Cron.Location) > 0 {
		endpoints["location"] = []string{trigger.Spec.Cron.Location}
	}
	if trigger.Spec.Cron.Args != nil {
		endpoints["args"] = []string{string(trigger.Spec.Cron.Args)}
	}
	if trigger.Spec.Cron.SaveLog {
		endpoints["saveLog"] = []string{fmt.Sprintf("%v", trigger.Spec.Cron.SaveLog)}
	}
	if trigger.Spec.Cron.SaveResult {
		endpoints["saveResult"] = []string{fmt.Sprintf("%v", trigger.Spec.Cron.SaveResult)}
	}
	return apis.EventSourceMappingConfiguration{
		EventSourceArn: fmt.Sprintf("arn:cron:%s", strings.TrimPrefix(trigger.Name, fmt.Sprintf("lambda-cron-%s-", trigger.Spec.FuncName))),
		FunctionArn:    trigger.Spec.FuncName,
		SelfManagedEventSource: apis.SelfManagedEventSource{
			Endpoints: endpoints,
		},
		UUID: trigger.Name,
	}, nil
}

func GetArnTriggerType(arn string) (string, error) {
	if arn == "http" {
		return HTTPTriggerType, nil
	}
	if arn == "cron" {
		return CronTriggerType, nil
	}
	return "", errors.New("not supported arn")
}
