package template

import "meta-egg/internal/domain/helper"

var TplGRPCBRTable = helper.PH_META_EGG_HEADER + `
package handler

import (
	"context"

	api "%%GO-MODULE%%/api/%%PROJECT-NAME-DIR%%"
	"%%GO-MODULE%%/internal/biz"
	"%%GO-MODULE%%/internal/common/cerror"
	"%%GO-MODULE%%/internal/common/contexts"
	"%%GO-MODULE%%/internal/repo/option"
	"google.golang.org/protobuf/types/known/emptypb"

	jgstr "github.com/Jinglever/go-string"
)

%%BR-RELATION-HANDLER-METHODS%%
`
