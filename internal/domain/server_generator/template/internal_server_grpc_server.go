package template

import "meta-egg/internal/domain/helper"

var TplInternalServerGRPCServer string = helper.PH_META_EGG_HEADER + `
package server

import (
	"context"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	log "%%GO-MODULE%%/pkg/log"
	api "%%GO-MODULE%%/api/%%PROJECT-NAME-DIR%%"
	"%%GO-MODULE%%/internal/common/resource"
	"google.golang.org/grpc"
	hdl "%%GO-MODULE%%/internal/handler/grpc"
)

// ProviderSet is grpc server providers.
var ProviderSet = wire.NewSet(
	NewServer,
)

type Config struct {
	Address string ` + "`" + `mapstructure:"address"` + "`" + ` // GRPC服务监听地址
	ReturnErrorDetail bool ` + "`" + `mapstructure:"return_error_detail"` + "`" + ` // http接口是否返回错误详情
	%%TPL-GRPC-SERVER-CONFIG-ACCESS-TOKEN%%}

type Server struct {
	Cfg      *Config
	Resource *resource.Resource
	Router   *gin.Engine
}

func NewServer(cfg *Config, rsrc *resource.Resource) *Server {
	s := &Server{
		Cfg:      cfg,
		Resource: rsrc,
	}
	return s
}

func (s *Server) Run() context.CancelFunc {
	// listen
	lis, err := net.Listen("tcp", s.Cfg.Address)
	if err != nil {
		log.WithError(err).Fatal("failed to listen")
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			errorInterceptor(s.Cfg),
			%%TPL-CALL-FUNC-GRPC-AUTH-INTERCEPTOR%%
		),
	)
	api.Register%%PROJECT-NAME-STRUCT%%Server(grpcServer, hdl.WireHandler(s.Resource))

	go func() {
		log.Infof("grpc server start at %s", s.Cfg.Address)
		err := grpcServer.Serve(lis)
		if err != nil {
			log.WithError(err).Fatal("grpc server error")
		}
	}()

	cancel := func() {
		log.Info("grpc server stop")
		grpcServer.GracefulStop()
	}
	return cancel
}
`

var TplCallFuncGRPCAuthInterceptor string = `authInterceptor(s.Resource.AccessToken, s.Cfg),`

var TplCallFuncGRPCExtractME string = `extractME(),`

var TplGRPCServerConfigAccessToken string = `VerifyAccessToken bool   ` + "`" + `mapstructure:"verify_access_token"` + "`" + ` // 是否验证access token, 为false时, 会仅解析JWT, 不会验证JWT签名
`
