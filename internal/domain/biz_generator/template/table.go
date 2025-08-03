package template

import "meta-egg/internal/domain/helper"

var TplCreateErrorDuplicateKey = `if errors.Is(err, gorm.ErrDuplicatedKey) {
		log.WithError(err).Error("fail to create %%TABLE-NAME%%, duplicated key")
		return cerror.AlreadyExists(err.Error())
	}
`

var TplUpdateErrorDuplicateKey = `if errors.Is(err, gorm.ErrDuplicatedKey) {
		log.WithError(err).Error("fail to update %%TABLE-NAME%%, duplicated key")
		return cerror.AlreadyExists(err.Error())
	}
`

var TplFuncCreate = `
func (b *BizService) Create%%TABLE-NAME-STRUCT%%(ctx context.Context, obj *%%TABLE-NAME-STRUCT%%BO) error {
	log := contexts.GetLogger(ctx).
		WithField("obj", jgstr.JsonEncode(obj))
	%%PREPARE-ASSIGN-BO-TO-MODEL%% m := &model.%%TABLE-NAME-STRUCT%%{%%ASSIGN-BO-TO-MODEL%%
	}
	return b.Resource.DB.Transaction(ctx, func(txCtx context.Context) error {
		err := b.%%TABLE-NAME-STRUCT%%Repo.Create(txCtx, m)
		if err != nil { %%CREATE-ERROR-DUPLICATE-KEY%%
			log.WithError(err).Error("fail to create %%TABLE-NAME%%")
			return cerror.Internal(err.Error())
		}
		%%RL-CREATE-IN-TRANSACTION%%
		bo, err := b.To%%TABLE-NAME-STRUCT%%BO(txCtx, m)
		if err != nil {
			log.WithError(err).Error("fail to convert %%TABLE-NAME%% model to %%TABLE-NAME-STRUCT%%BO")
			return cerror.Internal(err.Error())
		}
		*obj = *bo
		return nil
	})
}
`

var TplFuncGetList = `
%%RL-LIST-BO-DEFINITIONS%%type %%TABLE-NAME-STRUCT%%ListBO struct {%%COL-LIST-FOR-LIST%%
%%RL-LIST-BO-FIELDS%%}

func (b *BizService) To%%TABLE-NAME-STRUCT%%ListBO(ctx context.Context, ms []*model.%%TABLE-NAME-STRUCT%%) ([]*%%TABLE-NAME-STRUCT%%ListBO, error) {
	list := make([]*%%TABLE-NAME-STRUCT%%ListBO, 0, len(ms))
	for i := range ms {
		%%PREPARE-ASSIGN-MODEL-FOR-LIST%% list = append(list, &%%TABLE-NAME-STRUCT%%ListBO{%%ASSIGN-MODEL-FOR-LIST%%
%%RL-ASSIGN-MODEL-TO-LIST-BO%%		})
	}
	return list, nil
}

type %%TABLE-NAME-STRUCT%%FilterOption struct {
	%%FILTER-COL-LIST%%}

type %%TABLE-NAME-STRUCT%%ListOption struct {
	Pagination *option.PaginationOption
	Order      *option.OrderOption
	Filter     *%%TABLE-NAME-STRUCT%%FilterOption
}

func (b *BizService) Get%%TABLE-NAME-STRUCT%%List(ctx context.Context, opt *%%TABLE-NAME-STRUCT%%ListOption) ([]*%%TABLE-NAME-STRUCT%%ListBO, int64, error) {
	log := contexts.GetLogger(ctx).
		WithField("opt", jgstr.JsonEncode(opt))
	ms%%TABLE-NAME-STRUCT%%, total, err := b.%%TABLE-NAME-STRUCT%%Repo.GetList(ctx, &option.%%TABLE-NAME-STRUCT%%ListOption{
		Pagination: opt.Pagination,
		Order: opt.Order,%%ASSIGN-FILTER-TO-OPTION%%
		Select: []interface{}{%%COL-LIST-TO-SELECT-FOR-LIST%%
		},
	})
	if err != nil {
		log.WithError(err).Error("fail to get %%TABLE-NAME%% list")
		return nil, 0, cerror.Internal(err.Error())
	}
	list, err := b.To%%TABLE-NAME-STRUCT%%ListBO(ctx, ms%%TABLE-NAME-STRUCT%%)
	if err != nil {
		log.WithError(err).Error("fail to convert %%TABLE-NAME%% model to %%TABLE-NAME-STRUCT%%ListBO")
		return nil, 0, cerror.Internal(err.Error())
	}
	return list, total, nil
}
`

var TplFuncUpdate = `
type %%TABLE-NAME-STRUCT%%SetOption struct { %%SET-COL-LIST%%
}

func (b *BizService) Update%%TABLE-NAME-STRUCT%%ByID(ctx context.Context, id uint64, setOpt *%%TABLE-NAME-STRUCT%%SetOption) error {
	log := contexts.GetLogger(ctx).
		WithField("id", id).
		WithField("setOpt", jgstr.JsonEncode(setOpt))
	// assemble setCVs
	setCVs := make(map[string]interface{})
	%%SET-UPDATE-SETCVS%%
	if len(setCVs) == 0 {
		return nil
	}
	_, err := b.%%TABLE-NAME-STRUCT%%Repo.UpdateByID(ctx, id, setCVs, nil)
	if err != nil { %%UPDATE-ERROR-DUPLICATE-KEY%%
		log.WithError(err).Error("fail to update %%TABLE-NAME%%")
		return cerror.Internal(err.Error())
	}
	return nil
}
`

var TplFuncDelete = `
func (b *BizService) Delete%%TABLE-NAME-STRUCT%%ByID(ctx context.Context, id uint64) error {
	log := contexts.GetLogger(ctx).
		WithField("id", id)
	return b.Resource.DB.Transaction(ctx, func(txCtx context.Context) error {
		var err error
		%%RL-CASCADE-DELETE-IN-BIZ%%
		_, err = b.%%TABLE-NAME-STRUCT%%Repo.DeleteByID(txCtx, id)
		if err != nil {
			log.WithError(err).Error("fail to delete %%TABLE-NAME%%")
			return cerror.Internal(err.Error())
		}
		return nil
	})
}
`

var TplTable = helper.PH_META_EGG_HEADER + `
package biz

import (
	"context"
	"time"

	"%%GO-MODULE%%/gen/model"
	"%%GO-MODULE%%/pkg/gormx"
	"%%GO-MODULE%%/internal/common/cerror"
	"%%GO-MODULE%%/internal/repo/option"
	"%%GO-MODULE%%/internal/common/contexts"
	"gorm.io/gorm"
	"errors"

	jgstr "github.com/Jinglever/go-string"
	log "%%GO-MODULE%%/pkg/log"
)

%%RL-BO-DEFINITIONS%%

type %%TABLE-NAME-STRUCT%%BO struct {%%COL-LIST-IN-BO%%
%%RL-BO-FIELDS%%}

func (b *BizService) To%%TABLE-NAME-STRUCT%%BO(ctx context.Context, m *model.%%TABLE-NAME-STRUCT%%) (*%%TABLE-NAME-STRUCT%%BO, error) {
	%%PREPARE-ASSIGN-MODEL-TO-BO%% return &%%TABLE-NAME-STRUCT%%BO{%%ASSIGN-MODEL-TO-BO%%
%%RL-ASSIGN-MODEL-TO-BO%%	}, nil
}

%%TPL-FUNC-CREATE%%

func (b *BizService) Get%%TABLE-NAME-STRUCT%%ByID(ctx context.Context, id uint64) (*%%TABLE-NAME-STRUCT%%BO, error) {
	log := contexts.GetLogger(ctx).
		WithField("id", id)
	m%%TABLE-NAME-STRUCT%%, err := b.%%TABLE-NAME-STRUCT%%Repo.GetByID(ctx, id)
	if err != nil {
		log.WithError(err).Error("fail to get %%TABLE-NAME%% by id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cerror.NotFound(err.Error())
		} else {
			return nil, cerror.Internal(err.Error())
		}
	}
	%%TABLE-NAME-VAR%%BO, err := b.To%%TABLE-NAME-STRUCT%%BO(ctx, m%%TABLE-NAME-STRUCT%%)
	if err != nil {
		log.WithError(err).Error("fail to convert %%TABLE-NAME%% model to %%TABLE-NAME-STRUCT%%BO")
		return nil, cerror.Internal(err.Error())
	}
	return %%TABLE-NAME-VAR%%BO, nil
}

%%TPL-FUNC-GET-LIST%%
%%TPL-FUNC-UPDATE%%
%%TPL-FUNC-DELETE%%
%%RL-METHODS%%
%%BR-OPTIONS%%
%%BR-METHODS%%
`
