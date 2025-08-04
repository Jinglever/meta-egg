package template

import "meta-egg/internal/domain/helper"

var TplBRTable = helper.PH_META_EGG_HEADER + `
package biz

import (
	"context"

	"%%GO-MODULE%%/gen/model"
	"%%GO-MODULE%%/internal/common/cerror"
	"%%GO-MODULE%%/internal/common/contexts"
	"%%GO-MODULE%%/internal/repo/option"

	jgstr "github.com/Jinglever/go-string"
)

%%BR-RELATION-METHODS%%
`
