package template

import "meta-egg/internal/domain/helper"

var TplInternalOptionOption = helper.PH_META_EGG_HEADER + `
package option
// biz层的公共option

import "%%GO-MODULE%%/pkg/gormx"

const (
	OrderTypeAsc  = "asc"
	OrderTypeDesc = "desc"
)

const (
	DefaultPage = 1
	DefaultSize = 40
)

type PaginationOption struct {
	Page     int ` + "`" + `json:"page" form:"page"` + "`" + `      // 页码, 默认为1
	PageSize int ` + "`" + `json:"page_size" form:"size"` + "`" + ` // 每页数量, 默认为40
}

func (o *PaginationOption) GetOffset() int {
	if o.Page <= 0 {
		o.Page = DefaultPage
	}
	if o.PageSize <= 0 {
		o.PageSize = DefaultSize
	}
	return (o.Page - 1) * o.PageSize
}

func (o *PaginationOption) GetLimit() int {
	if o.PageSize <= 0 {
		o.PageSize = DefaultSize
	}
	return o.PageSize
}

func (o *PaginationOption) GetRepoOptions() []gormx.Option {
	ops := make([]gormx.Option, 0)
	ops = append(ops, gormx.Offset(int(o.GetOffset())))
	ops = append(ops, gormx.Limit(int(o.GetLimit())))
	return ops
}

type OrderOption struct {
	OrderBy   *string    ` + "`" + `json:"order_by" form:"order_by" binding:"omitempty"` + "`" + `     // 排序字段
	OrderType *string ` + "`" + `json:"order_type" form:"order_type" binding:"omitempty"` + "`" + ` // 排序类型,默认desc
}

func (o *OrderOption) GetRepoOptions(validOrderBy []string) []gormx.Option {
	ops := make([]gormx.Option, 0)
	if o.OrderBy == nil {
		return ops
	}
	if o.OrderType == nil {
		o.OrderType = new(string)
		*o.OrderType = OrderTypeDesc // 默认desc
	} else if *o.OrderType != OrderTypeAsc && *o.OrderType != OrderTypeDesc {
		*o.OrderType = OrderTypeDesc // 默认desc
	}
	for _, col := range validOrderBy {
		if col == *o.OrderBy {
			ops = append(ops, gormx.Order(col+" "+string(*o.OrderType)))
			break
		}
	}
	return ops
}
`
