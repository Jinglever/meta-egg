package template

import "meta-egg/internal/domain/helper"

var TplResourceStructDB string = `DB  gormx.DB
`

var TplResourceConfigStructDB string = `DB  *gormx.Config ` + "`" + `mapstructure:"db"` + "`" + `  // 数据库配置
`

var TplResourceDB string = `if cfg.DB != nil {
	db, err := gormx.ConnectDB(cfg.DB)
	if err != nil {
		log.WithError(err).Error("connect db failed, err")
		return nil, err
	}
	rsrc.DB = db
	rsrc.cancelFuncs = append(rsrc.cancelFuncs, func() {
		if err = rsrc.DB.Close(); err != nil {
			log.WithError(err).Errorf("close db error")
		}
		log.Info("close db success")
	})
}
`

var TplResourceStructAccessToken string = `AccessToken *jgjwt.JWT // access token jwt
`

var TplResourceConfigStructAccessToken string = `AccessToken *jgjwt.Config ` + "`" + `mapstructure:"access_token"` + "`" + ` // access token jwt配置
`

var TplResourceAccessToken string = `if cfg.AccessToken != nil {
	jwt, err := jgjwt.NewJWT(cfg.AccessToken)
	if err != nil {
		log.WithError(err).Error("init access token jwt failed")
		return nil, err
	}
	rsrc.AccessToken = jwt
}
`

var TplInternalCommonResourceResource string = helper.PH_META_EGG_HEADER + `
package resource

import (
	"context"

	jgjwt "github.com/Jinglever/go-jwt"
	log "%%GO-MODULE%%/pkg/log"
	"%%GO-MODULE%%/internal/common/contexts"
	"%%GO-MODULE%%/pkg/gormx"
	"gorm.io/gorm"
)

type Resource struct {
	cancelFuncs []context.CancelFunc
%%TPL-RESOURCE-STRUCT-DB%% %%TPL-RESOURCE-STRUCT-ACCESS-TOKEN%%}

type Config struct {
%%TPL-RESOURCE-CONFIG-STRUCT-DB%% %%TPL-RESOURCE-CONFIG-STRUCT-ACCESS-TOKEN%%}

func InitResource(cfg *Config) (*Resource, error) {
	rsrc := &Resource{
		cancelFuncs: make([]context.CancelFunc, 0),
	}

	%%TPL-RESOURCE-DB%%
	%%TPL-RESOURCE-ACCESS-TOKEN%%
	// your resource init here

	return rsrc, nil
}

func (r *Resource) Release() {
	for _, cancel := range r.cancelFuncs {
		cancel()
	}
}
`
