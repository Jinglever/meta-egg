package template

import "meta-egg/internal/domain/helper"

var TplInternalServerMonitorServer string = helper.PH_META_EGG_HEADER + `
package monitor

import (
	"context"
	"%%GO-MODULE%%/pkg/log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/google/wire"
)

// ProviderSet is monitor server providers.
var ProviderSet = wire.NewSet(
	NewServer,
)

type Config struct {
	Address string ` + "`" + `mapstructure:"address"` + "`" + ` // HTTP服务监听地址
}

type Server struct {
	Cfg    *Config
}

func NewServer(cfg *Config) *Server {
	s := &Server{
		Cfg: cfg,
	}
	return s
}

func (s *Server) Run() context.CancelFunc {
	httpServer := &http.Server{
		Addr:    s.Cfg.Address,
	}

	go func() {
		log.Info("monitor server start at ", s.Cfg.Address)
		err := httpServer.ListenAndServe()
		if err != http.ErrServerClosed {
			log.WithError(err).Fatalf("http server error, want %v", http.ErrServerClosed)
		}
	}()

	cancel := func() {
		log.Info("monitor server stop")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.WithError(err).Fatal("monitor server shutdown error")
		}
	}
	return cancel
}
`
