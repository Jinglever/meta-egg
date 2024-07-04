package template

import "meta-egg/internal/domain/helper"

var TplSetupMEForCreate string = `if me, ok := contexts.GetME(ctx); ok {
		%%SET-ME-FOR-CREATE%%
	}
`

var TplSetupMEForCreateBatch string = `if me, ok := contexts.GetME(ctx); ok {
		for _, m := range ms {
			%%SET-ME-FOR-CREATE%%
		}
	}
`

var TplSetupMEForUpdate string = `if me, ok := contexts.GetME(ctx); ok {
		setCVs[model.Col%%TABLE-NAME-STRUCT%%%%COL-UPDATED-BY%%] = me.ID
	}
`

var TplDefaultDelete string = `result := tx.Delete(&model.%%TABLE-NAME-STRUCT%%{})`
var TplSoftDeleteWithDeletedBy string = `var result *gorm.DB
	if me, ok := contexts.GetME(ctx); ok {
		result = tx.UpdateColumns(map[string]interface{}{
			model.Col%%TABLE-NAME-STRUCT%%%%COL-DELETED-BY%%: &(me.ID),
			model.Col%%TABLE-NAME-STRUCT%%%%COL-DELETED-AT%%: %%VAL-DELETED-AT%%,
		})
	} else {
		result = tx.Delete(&model.%%TABLE-NAME-STRUCT%%{})
	}`

var TplGenRepoDataTable = helper.PH_META_EGG_HEADER + `
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
	Gets(ctx context.Context, opts ...gormx.Option) ([]*model.%%TABLE-NAME-STRUCT%%, error)
	GetByID(ctx context.Context, id uint64, opts ...gormx.Option) (*model.%%TABLE-NAME-STRUCT%%, error)
	GetByIDs(ctx context.Context, ids []uint64, opts ...gormx.Option) ([]*model.%%TABLE-NAME-STRUCT%%, error)
	Create(ctx context.Context, m *model.%%TABLE-NAME-STRUCT%%) error
	CreateBatch(ctx context.Context, ms []*model.%%TABLE-NAME-STRUCT%%) error
	Update(ctx context.Context, setCVs map[string]interface{}, incCVs map[string]interface{}, opts ...gormx.Option) (rowsAffected int64, err error)
	UpdateByID(ctx context.Context, id uint64, setCVs map[string]interface{}, incCVs map[string]interface{}) (rowsAffected int64, err error)
	UpdateByIDs(ctx context.Context, ids []uint64, setCVs map[string]interface{}, incCVs map[string]interface{}) (rowsAffected int64, err error)
	Delete(ctx context.Context, opts ...gormx.Option) (rowsAffected int64, err error)
	DeleteByID(ctx context.Context, id uint64) (rowsAffected int64, err error)
	DeleteByIDs(ctx context.Context, ids []uint64) (rowsAffected int64, err error)
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

// create single record
func (s *%%TABLE-NAME-STRUCT%%RepoImpl) Create(ctx context.Context, m *model.%%TABLE-NAME-STRUCT%%) error {
%%TPL-SETUP-ME-FOR-CREATE%% return s.GetTX(ctx).Create(m).Error
}

// create batch
func (s *%%TABLE-NAME-STRUCT%%RepoImpl) CreateBatch(ctx context.Context, ms []*model.%%TABLE-NAME-STRUCT%%) error {
	if len(ms) == 0 {
		return nil
	}
	%%TPL-SETUP-ME-FOR-CREATE-BATCH%%
	return s.GetTX(ctx).CreateInBatches(ms, CreateBatchNum).Error
}

func (s *%%TABLE-NAME-STRUCT%%RepoImpl) Update(ctx context.Context, setCVs map[string]interface{}, incCVs map[string]interface{}, opts ...gormx.Option) (rowsAffected int64, err error) {
	tx := s.GetTX(ctx).Model(&model.%%TABLE-NAME-STRUCT%%{})
	for _, opt := range opts {
		tx = opt(tx)
	}
	if setCVs == nil {
		setCVs = make(map[string]interface{})
	}
	for k, v := range incCVs {
		setCVs[k] = gorm.Expr(k+" + ?", v)
	}
	if len(setCVs) == 0 {
		return 0, nil
	}
%%TPL-SETUP-ME-FOR-UPDATE%% result := tx.Updates(setCVs)
	if result.Error != nil {
		err = result.Error
	} else {
		rowsAffected = result.RowsAffected
	}
	return
}

// update by primary key
func (s *%%TABLE-NAME-STRUCT%%RepoImpl) UpdateByID(ctx context.Context, id uint64, setCVs map[string]interface{}, incCVs map[string]interface{}) (rowsAffected int64, err error) {
	return s.Update(ctx, setCVs, incCVs, gormx.Where(model.Col%%TABLE-NAME-STRUCT%%%%COL-ID%%+" = ?", id))
}

// update by primary keys
func (s *%%TABLE-NAME-STRUCT%%RepoImpl) UpdateByIDs(ctx context.Context, ids []uint64, setCVs map[string]interface{}, incCVs map[string]interface{}) (rowsAffected int64, err error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return s.Update(ctx, setCVs, incCVs, gormx.Where(model.Col%%TABLE-NAME-STRUCT%%%%COL-ID%%+" in (?)", ids))
}

func (s *%%TABLE-NAME-STRUCT%%RepoImpl) Delete(ctx context.Context, opts ...gormx.Option) (rowsAffected int64, err error) {
	tx := s.GetTX(ctx).Model(&model.%%TABLE-NAME-STRUCT%%{})
	for _, opt := range opts {
		tx = opt(tx)
	}
	%%TPL-DELETE%%
	if result.Error != nil {
		err = result.Error
	} else {
		rowsAffected = result.RowsAffected
	}
	return
}

// delete by primary key
func (s *%%TABLE-NAME-STRUCT%%RepoImpl) DeleteByID(ctx context.Context, id uint64) (rowsAffected int64, err error) {
	return s.Delete(ctx, gormx.Where(model.Col%%TABLE-NAME-STRUCT%%%%COL-ID%%+" = ?", id))
}

// delete by primary keys
func (s *%%TABLE-NAME-STRUCT%%RepoImpl) DeleteByIDs(ctx context.Context, ids []uint64) (rowsAffected int64, err error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return s.Delete(ctx, gormx.Where(model.Col%%TABLE-NAME-STRUCT%%%%COL-ID%%+" in (?)", ids))
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
