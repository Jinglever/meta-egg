package template

import "meta-egg/internal/domain/helper"

var TplMainRunHttpServer string = `// run http server
	gin.SetMode(gin.ReleaseMode)
	httpCancel := httpserver.NewServer(config.GetHTTPServerConfig(), rsrc).Run()
`

var TplMainRunGrpcServer string = `// run grpc server
	grpcCancel := grpcserver.NewServer(config.GetGRPCServerConfig(), rsrc).Run()
`

var TplMainCancelHttpServer string = `httpCancel()
`

var TplMainCancelGrpcServer string = `grpcCancel()
`

var TplMain = helper.PH_META_EGG_HEADER + `
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	log "%%GO-MODULE%%/pkg/log"
	"%%GO-MODULE%%/internal/common/resource"
	"%%GO-MODULE%%/internal/server/monitor"
	"%%GO-MODULE%%/internal/config"
	grpcserver "%%GO-MODULE%%/internal/server/grpc"
	httpserver "%%GO-MODULE%%/internal/server/http"
	"%%GO-MODULE%%/pkg/version"
)

var (
	configFilePath = flag.String("config", "./configs/%%PROJECT-NAME%%.yml", "config file")
)

func main() {
	flag.Parse()
	if *version.PrintVersion {
		version.Printer()
		return
	}

	// load config
	if err := config.LoadConfig(*configFilePath); err != nil {
		log.Fatalf("load config error %v", err)
	}

	// init log
	log.SetLevel(config.GetLogConfig().Level)

	// init resource
	rsrc, err := resource.InitResource(config.GetResourceConfig())
	if err != nil {
		log.Fatalf("init resource error %v", err)
	}

%%MAIN-RUN-HTTP-SERVER%% %%MAIN-RUN-GRPC-SERVER%%

	// run monitor server
	if config.GetMonitorServerConfig() != nil && config.GetMonitorServerConfig().Address != "" {
		monitorCancel := monitor.NewServer(config.GetMonitorServerConfig()).Run()
		defer monitorCancel()
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc
%%MAIN-CANCEL-HTTP-SERVER%% %%MAIN-CANCEL-GRPC-SERVER%% rsrc.Release()
	log.Info("%%PROJECT-NAME%% exit")
}
`
