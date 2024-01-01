package template

import "meta-egg/internal/domain/helper"

var TplBase = helper.PH_META_EGG_HEADER + `
package biz

import (
	"github.com/google/wire"
	"%%GO-MODULE%%/internal/common/resource"
	repo "%%GO-MODULE%%/internal/repo"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewBizService,
)

// 对数据实体的带业务规则的操作
type BizService struct {
	Resource         *resource.Resource
	%%REPO-LIST-IN-STRUCT%%
}


func NewBizService(
	rsrc *resource.Resource,
	%%REPO-LIST-IN-ARG%%) *BizService {
	return &BizService{
		Resource:   rsrc,
		%%ASSIGN-REPO-LIST%%
	}
}
`
