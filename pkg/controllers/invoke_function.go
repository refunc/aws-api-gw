package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	nats "github.com/nats-io/nats.go"
	"github.com/refunc/aws-api-gw/pkg/utils"
	"github.com/refunc/aws-api-gw/pkg/utils/awsutils"
	rfv1beta3 "github.com/refunc/refunc/pkg/apis/refunc/v1beta3"
	"github.com/refunc/refunc/pkg/client"
	"github.com/refunc/refunc/pkg/messages"
	rfutils "github.com/refunc/refunc/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

func InvokeFunction(c *gin.Context) {
	functionName := c.Param("FunctionName")
	//TODO support get function Qualifier
	var args json.RawMessage
	if err := c.BindJSON(&args); err != nil {
		klog.Error(err)
		awsutils.AWSErrorResponse(c, 400, "InvalidParameterValueException")
		return
	}

	funcdefLister, err := utils.GetFuncdefLister(c)
	if err != nil {
		klog.Error(err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	natsConn, err := utils.GetNatsConn(c)
	if err != nil {
		klog.Error(err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}

	region := c.GetString("region")
	fndef, err := funcdefLister.Funcdeves(region).Get(functionName)
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("get funcdef error %v", err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	if errors.IsNotFound(err) {
		awsutils.AWSErrorResponse(c, 404, "ResourceNotFoundException")
		return
	}

	invocationType := c.GetHeader(HeaderAmzInvocationType)
	if invocationType == "" {
		invocationType = "RequestResponse"
	}
	if invocationType == "RequestResponse" {
		logType := c.GetHeader(HeaderAmzLogType)
		if logType != "Tail" {
			logType = "None"
		}
		invokeRequestResponse(c, natsConn, args, logType, fndef)
	} else if invocationType == "Event" {
		invokeEvent(c, natsConn, args, fndef)
	} else if invocationType == "DryRun" {
		c.AbortWithStatus(204)
	} else {
		awsutils.AWSErrorResponse(c, 400, "InvalidParameterValueException")
	}
}

func invokeRequestResponse(c *gin.Context, natsConn *nats.Conn, args json.RawMessage, logType string, fndef *rfv1beta3.Funcdef) {
	request := &messages.InvokeRequest{
		Args:      args,
		RequestID: rfutils.GenID(args),
	}
	endpoint := fndef.Namespace + "/" + fndef.Name

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = client.WithLogger(ctx, klog.V(1))
	ctx = client.WithNatsConn(ctx, natsConn)
	ctx = client.WithTimeoutHint(ctx, time.Duration(fndef.Spec.Runtime.Timeout)*time.Second)
	ctx = client.WithLoggingHint(ctx, true)
	taskr, err := client.NewTaskResolver(ctx, endpoint, request)
	if err != nil {
		klog.Error(err)
		awsutils.AWSErrorResponse(c, 500, "ServiceException")
		return
	}
	logStream := taskr.LogObserver()
	var logs []byte
	for {
		select {
		case <-logStream.Changes():
			for logStream.HasNext() {
				logs = append(logs, []byte(logStream.Next().(string))...)
				logs = append(logs, messages.TokenCRLF...)
			}
		case <-taskr.Done():
			bts, err := taskr.Result()
			if err != nil {
				bts = messages.GetErrActionBytes(err)
			}
			c.Status(200)
			c.Header(HeaderAmzExecutedVersion, LambdaVersion)
			if logType == "Tail" {
				c.Header(HeaderAmzLogResult, string(logs))
			}
			// append \n split func result and aws api cli echo
			if !bytes.HasSuffix(bts, []byte{'\n'}) {
				bts = append(bts, '\n')
			}
			c.Writer.Write(bts)
			return
		}
	}
}

func invokeEvent(c *gin.Context, natsConn *nats.Conn, args json.RawMessage, fndef *rfv1beta3.Funcdef) {
	// TODO support asynchronously invoke
	awsutils.AWSErrorResponse(c, 400, "InvalidParameterValueException")
}
