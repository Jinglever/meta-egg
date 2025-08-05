package template

import "meta-egg/internal/domain/helper"

var TplInternalRepoTableData = helper.PH_META_EGG_HEADER + `
package repo

import (
	"context"
	"%%GO-MODULE%%/gen/model"
	gen "%%GO-MODULE%%/gen/repo"
	"%%GO-MODULE%%/internal/common/contexts"
	"%%GO-MODULE%%/internal/common/resource"
	"%%GO-MODULE%%/internal/repo/option"
	"%%GO-MODULE%%/pkg/gormx"
	jgstr "github.com/Jinglever/go-string"
)

//go:generate mockgen -package mock -destination ./mock/%%TABLE-NAME%%.go . %%TABLE-NAME-STRUCT%%Repo
type %%TABLE-NAME-STRUCT%%Repo interface {
	gen.%%TABLE-NAME-STRUCT%%Repo

	GetList(ctx context.Context, opt *option.%%TABLE-NAME-STRUCT%%ListOption) ([]*model.%%TABLE-NAME-STRUCT%%, int64, error)
	%%BR-REPO-INTERFACE%%
}

type %%TABLE-NAME-STRUCT%%RepoImpl struct {
	gen.%%TABLE-NAME-STRUCT%%RepoImpl
}

func New%%TABLE-NAME-STRUCT%%Repo(rsrc *resource.Resource) %%TABLE-NAME-STRUCT%%Repo {
	return &%%TABLE-NAME-STRUCT%%RepoImpl{
		%%TABLE-NAME-STRUCT%%RepoImpl: gen.%%TABLE-NAME-STRUCT%%RepoImpl{
			Resource: rsrc,
		},
	}
}

func (r *%%TABLE-NAME-STRUCT%%RepoImpl) GetList(ctx context.Context, opt *option.%%TABLE-NAME-STRUCT%%ListOption) ([]*model.%%TABLE-NAME-STRUCT%%, int64, error) {
	log := contexts.GetLogger(ctx).
		WithField("opt", jgstr.JsonEncode(opt))
	ops := make([]gormx.Option, 0)
	if opt != nil && opt.Filter != nil {
		ops = append(ops, opt.Filter.GetRepoOptions()...)
	}
	if opt != nil && opt.Order != nil {
		validOrderby := []string{
			%%ORDER-BY-LIST%%
		}
		ops = append(ops, opt.Order.GetRepoOptions(validOrderby)...)
	}
	total, err := r.Count(ctx, ops...)
	if err != nil {
		log.WithError(err).Error("fail to count %%TABLE-NAME%% list")
		return nil, 0, err
	}
	if total == 0 {
		return make([]*model.%%TABLE-NAME-STRUCT%%, 0), 0, nil
	}
	if opt != nil && opt.Pagination != nil {
		ops = append(ops, opt.Pagination.GetRepoOptions()...)
	}
	if opt != nil && len(opt.Select) > 0 {
		ops = append(ops, gormx.Select(opt.Select...))
	}
	ms, err := r.Gets(ctx, ops...)
	if err != nil {
		log.WithError(err).Error("fail to get %%TABLE-NAME%% list")
		return nil, 0, err
	}
	return ms, total, nil
}

%%BR-REPO-IMPLEMENTATION%%
`
