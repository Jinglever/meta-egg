package template

import "meta-egg/internal/domain/helper"

var TplPkgGormxDb string = helper.PH_META_EGG_HEADER + `
package gormx

import (
	"context"

	"github.com/google/wire"
	"gorm.io/gorm"
)

//go:generate mockgen -package mock -destination ./mock/db.go . DB
type DB interface {
	GetTX(ctx context.Context) *gorm.DB
	SetTX(ctx context.Context, tx *gorm.DB) context.Context
	Close() error
	Transaction(ctx context.Context, fc func(txCtx context.Context) error) error
}

type DBImpl struct {
	DB *gorm.DB
}

func (d *DBImpl) GetTX(ctx context.Context) *gorm.DB {
	tx, ok := getTX(ctx)
	if !ok {
		tx = d.DB
	}
	return tx
}

func (d *DBImpl) SetTX(ctx context.Context, tx *gorm.DB) context.Context {
	if tx == nil {
		return setTX(ctx, d.DB)
	} else {
		return setTX(ctx, tx)
	}
}

func (d *DBImpl) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *DBImpl) Transaction(ctx context.Context, fc func(txCtx context.Context) error) error {
	tx := d.GetTX(ctx)
	return tx.Transaction(func(newTx *gorm.DB) error {
		txCtx := d.SetTX(ctx, newTx) // set tx to context
		return fc(txCtx)
	})
}
`
