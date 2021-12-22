package controllers

import (
	"encoding/json"
	"time"

	"github.com/refunc/aws-api-gw/pkg/apis"
	rfv1beta3 "github.com/refunc/refunc/pkg/apis/refunc/v1beta3"
)

const (
	LambdaVersion          = "0" //TODO support lambda version apis
	LambdaLabelName        = "lambda.refunc.io/name"
	LambdaLabelVersion     = "lambda.refunc.io/version"
	LambdaLabelAutoCreated = "lambda.refunc.io/auto-created"
	TriggerType            = "httptrigger"
)

func FuncdefToLambdaConfiguration(fndef rfv1beta3.Funcdef) (apis.FunctionConfiguration, error) {
	custom := map[string]interface{}{}
	err := json.Unmarshal(fndef.Spec.Custom, &custom)
	if err != nil {
		return apis.FunctionConfiguration{}, err
	}
	codeSize := int64(custom["codeSize"].(float64))
	return apis.FunctionConfiguration{
		CodeSha256: fndef.Spec.Hash,
		CodeSize:   codeSize,
		Environment: apis.FunctionEnvironment{
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
