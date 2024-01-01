package template

import "meta-egg/internal/domain/helper"

var TplConfigStructHttpServer string = `
HTTPServer *httpserver.Config ` + "`" + `mapstructure:"http_server"` + "`" + ` // http服务配置`

var TplConfigStructGrpcServer string = `
GRPCServer *grpcserver.Config ` + "`" + `mapstructure:"grpc_server"` + "`" + ` // grpc服务配置`

var TplConfigFuncGetHttpServer string = `
func GetHTTPServerConfig() *httpserver.Config {
	return cfg.HTTPServer
}
`

var TplConfigFuncGetGrpcServer string = `
func GetGRPCServerConfig() *grpcserver.Config {
	return cfg.GRPCServer
}
`

var TplInternalConfigConfig = helper.PH_META_EGG_HEADER + `
package config

import (
	"time"
	jgconf "github.com/Jinglever/go-config"
	"github.com/Jinglever/go-config/option"
	"%%GO-MODULE%%/internal/common/constraint"
	"%%GO-MODULE%%/internal/common/resource"
	grpcserver "%%GO-MODULE%%/internal/server/grpc"
	httpserver "%%GO-MODULE%%/internal/server/http"
	"%%GO-MODULE%%/internal/server/monitor"
	log "%%GO-MODULE%%/pkg/log"
	"%%GO-MODULE%%/internal/common/contexts"
)

type Config struct {
	Log        *LogConfig         ` + "`" + `mapstructure:"log"` + "`" + `         // 系统日志配置
	Constraint *constraint.Config ` + "`" + `mapstructure:"constraint"` + "`" + `  // 业务约束配置
	Resource   *resource.Config   ` + "`" + `mapstructure:"resource"` + "`" + `    // 资源配置 %%TPL-CONFIG-STRUCT-HTTP-SERVER%% %%TPL-CONFIG-STRUCT-GRPC-SERVER%%	
	MonitorServer *monitor.Config    ` + "`" + `mapstructure:"monitor_server"` + "`" + ` // 监控服务配置
}

type LogConfig struct {
	Level string ` + "`" + `mapstructure:"level"` + "`" + ` // 日志级别 debug, info, warn, error, fatal
}

var cfg *Config

const EnvPrefix = "%%ENV-PREFIX%%"

func LoadConfig(path string) error {
	var c Config
	if err := jgconf.LoadYamlConfig(path, &c, option.WithEnvPrefix(EnvPrefix)); err != nil {
		return err
	}
	cfg = &c

	// 业务约束配置
	if cfg.Constraint != nil {
		constraint.InjectConfig(*cfg.Constraint)

		// 设定时区
		loc, err := time.LoadLocation(constraint.GetConfig().Timezone)
		if err != nil {
			log.WithError(err).Error("load location failed")
			return err
		}
		time.Local = loc
	}
	return nil
}

func GetLogConfig() *LogConfig {
	return cfg.Log
}

func GetResourceConfig() *resource.Config {
	return cfg.Resource
}

%%TPL-CONFIG-FUNC-GET-HTTP-SERVER%% %%TPL-CONFIG-FUNC-GET-GRPC-SERVER%%

func GetMonitorServerConfig() *monitor.Config {
	return cfg.MonitorServer
}
`
