package template

import "meta-egg/internal/domain/helper"

var TplGRPCBase string = helper.PH_META_EGG_HEADER + `
package handler

import (
	"github.com/google/wire"
	"%%GO-MODULE%%/internal/common/resource"
	api "%%GO-MODULE%%/api/%%PROJECT-NAME-DIR%%"
	"%%GO-MODULE%%/internal/biz"
	%%IMPORT-USECASE-LIST%%
)

// ProviderSet is grpc handler providers.
var ProviderSet = wire.NewSet(
	NewHandler,
)

type Handler struct {
	api.Unimplemented%%PROJECT-NAME-STRUCT%%Server
	Resource      *resource.Resource
	BizService          *biz.BizService
	%%USECASE-LIST-IN-STRUCT%% // TODO: add your usecase
}

func NewHandler(
	rsrc *resource.Resource,
	bizService *biz.BizService,
	%%USECASE-LIST-IN-ARG%% // TODO: add your usecase
) *Handler {
	return &Handler{
		Resource:   rsrc,
		BizService:          bizService,
		%%ASSIGN-USECASE-LIST%% // TODO: setup your usecase
	}
}
`
