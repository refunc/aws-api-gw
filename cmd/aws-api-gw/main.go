package main

import (
	"context"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/Arvintian/go-utils/cmdutil"
	"github.com/Arvintian/go-utils/cmdutil/flagtools"
	"github.com/Arvintian/go-utils/cmdutil/pflagenv"
	"github.com/gin-gonic/gin"
	"github.com/refunc/aws-api-gw/pkg/routers"
	"github.com/refunc/aws-api-gw/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	_ "github.com/refunc/refunc/pkg/env"
	"github.com/refunc/refunc/pkg/utils/cmdutil/sharedcfg"
)

var config struct {
	routerCfg routers.Config
	Debug     bool
	Addr      string
	Namespace string
}

func main() {
	runtime.GOMAXPROCS(16 * runtime.NumCPU())
	rand.Seed(time.Now().UTC().UnixNano())

	klog.CopyStandardLogTo("INFO")
	defer klog.Flush()

	cmd := &cobra.Command{
		Use:   "aws-api-gw",
		Short: "Start refunc aws lambda api gateway.",
		Run: func(cmd *cobra.Command, args []string) {

			// gin setting
			gin.DisableConsoleColor()

			if config.Debug {
				gin.SetMode(gin.DebugMode)
			} else {
				gin.SetMode(gin.ReleaseMode)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sc := sharedcfg.New(ctx, config.Namespace)

			// create router and init informers
			router := routers.CreateHTTPRouter(sc.Configs(), config.routerCfg, ctx.Done())

			go func() {
				klog.Infof("Refunc aws lambda api gateway version: %s\n", version.Version)
				klog.Infof("Listening and serving HTTP on %s\n", config.Addr)

				srv := &http.Server{
					Addr:    config.Addr,
					Handler: router,
				}

				if err := srv.ListenAndServe(); err != nil {
					klog.Error(err)
				}
			}()

			go func() {
				// informers started
				sc.Run(ctx.Done())
			}()

			klog.Infof(`Received signal "%v", exiting...`, <-cmdutil.GetSysSig())

		},
	}

	cmd.Flags().StringVar(&config.Addr, "addr", "0.0.0.0:9000", "ListenAndServe Address.")
	cmd.Flags().BoolVar(&config.routerCfg.Rbac, "rbac", false, "Enable rbac auth.")
	cmd.Flags().BoolVar(&config.Debug, "debug", false, "Enable gin's debug mode.")
	cmd.Flags().StringVarP(&config.Namespace, "namespace", "n", "", "The scope of namepsace to manipulate.")
	flagtools.BindFlags(cmd.PersistentFlags())

	// set global flags using env
	pflagenv.ParseSet(pflagenv.Prefix, cmd.PersistentFlags())

	if err := cmd.Execute(); err != nil {
		klog.Fatal(err)
	}

}
