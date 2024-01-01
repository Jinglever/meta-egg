package template

import "meta-egg/internal/domain/helper"

var TplHTTPBase string = helper.PH_META_EGG_HEADER + `
package handler

import (
	"github.com/google/wire"
	"%%GO-MODULE%%/internal/common/resource"
	"github.com/gin-gonic/gin"
	api "%%GO-MODULE%%/api/%%PROJECT-NAME-DIR%%"
	"%%GO-MODULE%%/internal/common/cerror"
	"strings"
	"github.com/go-playground/validator/v10"
	log "%%GO-MODULE%%/pkg/log"
	"%%GO-MODULE%%/internal/common/contexts"
	"%%GO-MODULE%%/internal/biz"
	%%IMPORT-USECASE-LIST%%
)

// ProviderSet is http handler providers.
var ProviderSet = wire.NewSet(
	NewHandler,
)

type RspBase struct {
	Code    api.ErrCode ` + "`" + `json:"code"` + "`" + `
	Message string          ` + "`" + `json:"message"` + "`" + `
}

type RspData struct {
	RspBase
	Data interface{} ` + "`" + `json:"data"` + "`" + `
}

// data如果为nil, 则响应RspBase{}; 否则响应RspData{}
func ResponseSuccess(c *gin.Context, data interface{}) {
	e := cerror.Ok()
	if data != nil {
		c.JSON(
			e.HttpStatus,
			RspData{
				RspBase: RspBase{
					Code:    e.Code,
					Message: e.Error(),
				},
				Data: data,
			},
		)
	} else {
		c.JSON(
			e.HttpStatus,
			RspBase{
				Code:    e.Code,
				Message: e.Error(),
			},
		)
	}
}

func ResponseFail(c *gin.Context, err error) {
	var (
		cErr *cerror.CustomError
		ok   bool
	)
	if cErr, ok = err.(*cerror.CustomError); !ok {
		cErr = cerror.Unknown(err.Error())
	}
	c.JSON(
		cErr.HttpStatus,
		RspBase{
			Code:    cErr.Code,
			Message: cErr.Error(),
		},
	)
}

// shouldbind and validate
func shouldBind(c *gin.Context, req interface{}) error {
	log := contexts.GetLogger(c.Request.Context())
	if err := c.ShouldBind(req); err != nil {
		log.WithError(err).Error("c.ShouldBind failed")
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return cerror.InvalidArgument(err.Error())
		} else {
			fields := make([]string, 0, len(validationErrors))
			for _, err := range validationErrors {
				fields = append(fields, err.Field())
			}
			return cerror.InvalidArgument(strings.Join(fields, ", "))
		}
	}
	return nil
}

type Handler struct {
	Resource      *resource.Resource
	BizService          *biz.BizService
	%%USECASE-LIST-IN-STRUCT%% // TODO: add your usecase
}

func NewHandler(
	rsrc *resource.Resource,
	bizService *biz.BizService,
	%%USECASE-LIST-IN-ARG%% // TODO: add usecase
) *Handler {
	return &Handler{
		Resource:   rsrc,
		BizService:          bizService,
		%%ASSIGN-USECASE-LIST%% // TODO: setup usecase
	}
}
`
