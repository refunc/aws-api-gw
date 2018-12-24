package main

import (
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"os"
	"runtime"
	"time"

	nats "github.com/nats-io/go-nats"
	"github.com/refunc/refunc/pkg/client"
	"github.com/refunc/refunc/pkg/env"
	"github.com/refunc/refunc/pkg/messages"
	"github.com/refunc/refunc/pkg/utils"
	"github.com/refunc/refunc/pkg/utils/cmdutil"
	"github.com/refunc/refunc/pkg/utils/cmdutil/flagtools"
	"github.com/spf13/pflag"
	"k8s.io/klog"
)

var config struct {
	Namespace string
	Name      string
	Timeout   time.Duration
}

func init() {
	pflag.StringVarP(&config.Namespace, "namespace", "n", getEnviron("REFUNC_NAMESPACE", ""), "The namespace to lookup function")
	pflag.DurationVarP(&config.Timeout, "timeout", "t", messages.DefaultJobTimeout, "The timeout of this invocation in second")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UTC().UnixNano())

	flagtools.InitFlags()

	klog.CopyStandardLogTo("INFO")
	defer klog.Flush()

	if config.Namespace == "" {
		klog.Exit("namespace must be set")
	}
	if len(pflag.Args()) == 0 {
		klog.Exit("function name  must be passed as argument")
	}
	config.Name = pflag.Args()[0]
	endpoint := config.Namespace + "/" + config.Name

	natsConn, err := env.NewNatsConn(nats.Name(endpoint + ".invoke"))
	if err != nil {
		klog.Exitf("Failed to connect to nats %s, %v", env.GlobalNatsURLString(), err)
	}
	defer natsConn.Close()

	var args json.RawMessage
	if err := json.NewDecoder(os.Stdin).Decode(&args); err != nil {
		if err != io.EOF {
			klog.Error(err)
			return
		}
		args = []byte{'{', '}'}
	}
	request := &messages.InvokeRequest{
		Args:      args,
		RequestID: utils.GenID(args),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = client.WithLogger(ctx, klog.V(1))
	ctx = client.WithNatsConn(ctx, natsConn)
	ctx = client.WithTimeoutHint(ctx, config.Timeout)
	ctx = client.WithLoggingHint(ctx, true)
	taskr, err := client.NewTaskResolver(ctx, endpoint, request)
	if err != nil {
		klog.Error(err)
		return
	}

	sig := cmdutil.GetSysSig()

	select {
	case <-sig:
		klog.Infof(`received signal "%v", exiting...`, sig)
		return
	case <-taskr.Done():
		bts, err := taskr.Result()
		if err != nil {
			klog.Error(err)
			return
		}
		os.Stdout.Write(bts)
		os.Stdout.Sync()
		return
	}
}

func getEnviron(key, alter string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return alter
}
