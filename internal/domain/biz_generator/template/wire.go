package template

var TplInternalBizWire string = `//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package biz

import (
	"github.com/google/wire"
	"%%GO-MODULE%%/internal/common/resource"
	%%COMMENT-REPO%%repo "%%GO-MODULE%%/internal/repo"
)

func WireBizService(rsrc *resource.Resource) *BizService {
	panic(wire.Build(
		%%COMMENT-REPO%%repo.ProviderSet,
		ProviderSet,
	))
}
`
