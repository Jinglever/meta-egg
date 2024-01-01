package template

import "meta-egg/internal/domain/helper"

var TplUsecaseBase = helper.PH_META_EGG_HEADER + `
package %%USECASE-NAME-PKG%%

import (
	"%%GO-MODULE%%/internal/biz"
	"%%GO-MODULE%%/internal/common/resource"
)

// %%USECASE-DESC%%
type %%USECASE-NAME-STRUCT%%Usecase struct {
	Resource   *resource.Resource
	BizService *biz.BizService
}

func New%%USECASE-NAME-STRUCT%%Usecase(
	rsrc *resource.Resource,
	bizService *biz.BizService,
) *%%USECASE-NAME-STRUCT%%Usecase {
	return &%%USECASE-NAME-STRUCT%%Usecase{
		Resource:   rsrc,
		BizService: bizService,
	}
}

`
