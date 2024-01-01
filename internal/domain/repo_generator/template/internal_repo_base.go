package template

import "meta-egg/internal/domain/helper"

var TplInternalRepoBase string = helper.PH_META_EGG_HEADER + `
package repo

import (
	"github.com/google/wire"
	// mock "%%GO-MODULE%%/internal/repo/mock"
)

// ProviderSet is repo providers.
var ProviderSet = wire.NewSet(
	%%NEW-REPO-FUNC-LIST-IN-PROVIDER-SET%%
)

// MockProviderSet is mock repo providers.
var MockProviderSet = wire.NewSet(
	%%NEW-REPO-FUNC-LIST-IN-MOCK-PROVIDER-SET%%
)
`
