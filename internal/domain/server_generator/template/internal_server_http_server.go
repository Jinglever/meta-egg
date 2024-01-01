package template

import "meta-egg/internal/domain/helper"

var TplInternalServerHTTPServer string = helper.PH_META_EGG_HEADER + `
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	log "%%GO-MODULE%%/pkg/log"
	"%%GO-MODULE%%/internal/common/resource"
)

// ProviderSet is http server providers.
var ProviderSet = wire.NewSet(
	NewServer,
)

type Config struct {
	Address           string        ` + "`" + `mapstructure:"address"` + "`" + ` // HTTP服务监听地址
	ReadHeaderTimeout time.Duration ` + "`" + `mapstructure:"read_header_timeout"` + "`" + ` // HTTP服务读取请求头超时时间
	ReturnErrorDetail bool ` + "`" + `mapstructure:"return_error_detail"` + "`" + ` // http接口是否返回错误详情
	%%TPL-HTTP-SERVER-CONFIG-ACCESS-TOKEN%%}

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
	s.initRouter()
	return s
}

func (s *Server) Run() context.CancelFunc {
	httpServer := &http.Server{
		Addr:        s.Cfg.Address,
		Handler:     s.Router,
		ReadTimeout: s.Cfg.ReadHeaderTimeout,
	}
	go func() {
		log.Info("http server start at ", s.Cfg.Address)
		err := httpServer.ListenAndServe()
		if err != http.ErrServerClosed {
			log.WithError(err).Fatalf("http server error, want %v", http.ErrServerClosed)
		}
	}()

	cancel := func() {
		log.Info("http server stop")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.WithError(err).Fatal("http server shutdown error")
		}
	}
	return cancel
}
`

var TplHTTPServerConfigAccessToken string = `VerifyAccessToken         bool          ` + "`" + `mapstructure:"verify_access_token"` + "`" + `          // 是否验证access token, 为false时, 会仅解析JWT, 不会验证JWT签名
`
