package template

import "meta-egg/internal/domain/helper"

var TplBase = helper.PH_META_EGG_HEADER + `
package usecase

import (
	"github.com/google/wire"
	%%IMPORT-USECASE-LIST%%
)

// ProviderSet is domain providers.
var ProviderSet = wire.NewSet(
	%%PROVIDER-USECASE-LIST%%
)
`
