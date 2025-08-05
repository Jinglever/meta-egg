package template

import "meta-egg/internal/domain/helper"

var TplInternalOptionTable = helper.PH_META_EGG_HEADER + `
package option

import (
	"%%GO-MODULE%%/gen/model"
	"%%GO-MODULE%%/pkg/gormx"
)

type %%TABLE-NAME-STRUCT%%FilterOption struct {
%%FILTER-COL-LIST%%}

func (o *%%TABLE-NAME-STRUCT%%FilterOption) GetRepoOptions() []gormx.Option {
	ops := make([]gormx.Option, 0)
	%%FILTER-GET-REPO-OPTIONS%%return ops
}

type %%TABLE-NAME-STRUCT%%ListOption struct {
	Pagination *PaginationOption
	Order      *OrderOption
	Filter     *%%TABLE-NAME-STRUCT%%FilterOption
	Select     []interface{} // select columns, such as []interface{}{"id", "name"}
}

%%BR-OPTIONS%%
`
