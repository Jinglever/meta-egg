package template

import "meta-egg/internal/domain/helper"

var TplGenRepoMetaTable = helper.PH_META_EGG_HEADER + `
package repo

import (
	"context"
	"time"

	"%%GO-MODULE%%/gen/model"
	"%%GO-MODULE%%/pkg/gormx"
	"%%GO-MODULE%%/internal/common/contexts"
	"%%GO-MODULE%%/internal/common/resource"
	"gorm.io/gorm"
)

type %%TABLE-NAME-STRUCT%%Repo interface {
	GetTX(ctx context.Context) *gorm.DB
	GetSemanticByID(id uint64) string
	Gets(ctx context.Context, opts ...gormx.Option) ([]*model.%%TABLE-NAME-STRUCT%%, error)
	GetByID(ctx context.Context, id uint64, opts ...gormx.Option) (*model.%%TABLE-NAME-STRUCT%%, error)
	GetByIDs(ctx context.Context, ids []uint64, opts ...gormx.Option) ([]*model.%%TABLE-NAME-STRUCT%%, error)
	Count(ctx context.Context, opts ...gormx.Option) (count int64, err error)
}

type %%TABLE-NAME-STRUCT%%RepoImpl struct {
	Resource *resource.Resource
}

// get db
func (s *%%TABLE-NAME-STRUCT%%RepoImpl) GetTX(ctx context.Context) *gorm.DB {
	// in case of transaction
	return s.Resource.DB.GetTX(ctx)
}

func (s *%%TABLE-NAME-STRUCT%%RepoImpl) GetSemanticByID(id uint64) string {
	switch id {%%CASE-META-ID-TO-SEMANTIC%%
	default:
		return ""
	}
}

func (s *%%TABLE-NAME-STRUCT%%RepoImpl) Gets(ctx context.Context, opts ...gormx.Option) ([]*model.%%TABLE-NAME-STRUCT%%, error) {
	var ms []*model.%%TABLE-NAME-STRUCT%%
	tx := s.GetTX(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

// get by primary key
func (s *%%TABLE-NAME-STRUCT%%RepoImpl) GetByID(ctx context.Context, id uint64, opts ...gormx.Option) (*model.%%TABLE-NAME-STRUCT%%, error) { //nolint
	var ms []*model.%%TABLE-NAME-STRUCT%%
	opts = append(opts, gormx.Where(model.Col%%TABLE-NAME-STRUCT%%%%COL-ID%%+" = ?", id), gormx.Limit(1))
	ms, err := s.Gets(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(ms) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return ms[0], nil
}

// get by primary keys
func (s *%%TABLE-NAME-STRUCT%%RepoImpl) GetByIDs(ctx context.Context, ids []uint64, opts ...gormx.Option) ([]*model.%%TABLE-NAME-STRUCT%%, error) {
	if len(ids) == 0 {
		return make([]*model.%%TABLE-NAME-STRUCT%%, 0), nil
	}
	var ms []*model.%%TABLE-NAME-STRUCT%%
	opts = append(opts, gormx.Where(model.Col%%TABLE-NAME-STRUCT%%%%COL-ID%%+" in (?)", ids))
	ms, err := s.Gets(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return ms, nil
}

// count
func (s *%%TABLE-NAME-STRUCT%%RepoImpl) Count(ctx context.Context, opts ...gormx.Option) (count int64, err error) {
	tx := s.GetTX(ctx).Model(&model.%%TABLE-NAME-STRUCT%%{})
	for _, opt := range opts {
		tx = opt(tx)
	}
	result := tx.Count(&count)
	if result.Error != nil {
		err = result.Error
	}
	return
}
`
