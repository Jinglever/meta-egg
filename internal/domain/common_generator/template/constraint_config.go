package template

import "meta-egg/internal/domain/helper"

var TplConstraintConfig = helper.PH_META_EGG_HEADER + `
package constraint

var cfg *Config

// 通过配置文件获取配置
type Config struct {
	Timezone             string ` + "`" + `mapstructure:"timezone"` + "`" + `               // 时区 格式: Asia/Shanghai
}

var (
	defaultTimezone      = "Asia/Shanghai"
)

func InjectConfig(c Config) {
	cfg = &c
	if c.Timezone == "" {
		cfg.Timezone = defaultTimezone
	}
}

func GetConfig() Config {
	if cfg == nil {
		// 默认配置
		cfg = &Config{
			Timezone:             defaultTimezone,
		}
	}
	return *cfg
}
`
