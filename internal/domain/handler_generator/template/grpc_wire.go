package template

var TplInternalHandlerGRPCWire string = `//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package handler

import (
	"github.com/google/wire"
	"%%GO-MODULE%%/internal/common/resource"
	"%%GO-MODULE%%/internal/biz"
	%%COMMENT-REPO%%repo "%%GO-MODULE%%/internal/repo"
	%%COMMENT-DOMAIN%%"%%GO-MODULE%%/internal/usecase"
)

func WireHandler(rsrc *resource.Resource) *Handler {
	panic(wire.Build(
		%%COMMENT-REPO%%repo.ProviderSet,
		biz.ProviderSet,
		%%COMMENT-DOMAIN%%usecase.ProviderSet,
		ProviderSet,
	))
}
`
