package template

import "meta-egg/internal/domain/helper"

var TplPkgGormxTransaction string = helper.PH_META_EGG_HEADER + `
package gormx

import (
	"context"

	"gorm.io/gorm"
)

type CtxKeyTX struct{}

func setTX(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, CtxKeyTX{}, tx)
}

func getTX(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(CtxKeyTX{}).(*gorm.DB)
	return tx, ok
}
`
