package template

import "meta-egg/internal/domain/helper"

var TplHTTPBRTable = helper.PH_META_EGG_HEADER + `
package handler

import (
	"context"
	"%%GO-MODULE%%/internal/biz"
	"%%GO-MODULE%%/internal/common/constraint"
	"%%GO-MODULE%%/internal/common/contexts"
	"%%GO-MODULE%%/internal/repo/option"

	jgstr "github.com/Jinglever/go-string"
	"github.com/gin-gonic/gin"
)

%%BR-RELATION-HANDLER-METHODS%%
`
