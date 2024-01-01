package template

import "meta-egg/internal/domain/helper"

var TplHTTPHandlerCreate string = `type ReqCreate%%TABLE-NAME-STRUCT%% struct {%%COL-LIST-FOR-CREATE%%
}

//	@Id			Create%%TABLE-NAME-STRUCT%%
//	@Tags		%%TABLE-COMMENT%%
//	@Summary	创建%%TABLE-COMMENT%%
//	@Description
//	@Accept		json
//	@Produce	json
//	@Param		Authorization	header		string			true	"Bearer <jwt-token>"
//	@Param		body			body		ReqCreate%%TABLE-NAME-STRUCT%%	true	"%%TABLE-COMMENT%%"
//	@Success	200				{object}	RspData{data=%%TABLE-NAME-STRUCT%%Detail}
//	@Failure	400				{object}	RspBase
//	@Router		/api/v1/%%TABLE-NAME-URI%% [post]
func (h *Handler) Create%%TABLE-NAME-STRUCT%%(c *gin.Context) {
	var req ReqCreate%%TABLE-NAME-STRUCT%%
	err := shouldBind(c, &req)
	if err != nil {
		ResponseFail(c, err)
		return
	}
	ctx := c.Request.Context()
	log := contexts.GetLogger(ctx).
		WithField("req", jgstr.JsonEncode(req))

	%%PREPARE-ASSIGN-CREATE-TO-BO%% %%TABLE-NAME-VAR%%BO := &biz.%%TABLE-NAME-STRUCT%%BO{%%ASSIGN-CREATE-TO-BO%%
	}
	err = h.BizService.Create%%TABLE-NAME-STRUCT%%(ctx, %%TABLE-NAME-VAR%%BO)
	if err != nil {
		log.WithError(err).Error("BizService.Create%%TABLE-NAME-STRUCT%% failed")
		ResponseFail(c, err)
		return
	}
	d, err := h.To%%TABLE-NAME-STRUCT%%Detail(ctx, %%TABLE-NAME-VAR%%BO)
	if err != nil {
		log.WithError(err).Error("convert %%TABLE-NAME-STRUCT%%BO to %%TABLE-NAME-STRUCT%%Detail failed")
		ResponseFail(c, err)
		return
	}
	ResponseSuccess(c, d)
}
`

var TplHTTPHandlerGetList string = `// %%TABLE-COMMENT%%列表信息
type %%TABLE-NAME-STRUCT%%ListInfo struct {%%COL-LIST-FOR-LIST%%
}

func (h *Handler) To%%TABLE-NAME-STRUCT%%ListInfo(ctx context.Context, objs []*biz.%%TABLE-NAME-STRUCT%%ListBO) ([]*%%TABLE-NAME-STRUCT%%ListInfo, error) {
	list := make([]*%%TABLE-NAME-STRUCT%%ListInfo, 0, len(objs))
	for i := range objs {
		%%PREPARE-ASSIGN-BO-FOR-LIST%% list = append(list, &%%TABLE-NAME-STRUCT%%ListInfo{%%ASSIGN-BO-FOR-LIST%%
		})
	}
	return list, nil
}

// %%TABLE-COMMENT%%列表
type %%TABLE-NAME-STRUCT%%List struct {
	List  []*%%TABLE-NAME-STRUCT%%ListInfo ` + "`" + `json:"list"` + "`" + `  // %%TABLE-COMMENT%%列表
	Total int64         ` + "`" + `json:"total"` + "`" + ` // 总数
}

type ReqGet%%TABLE-NAME-STRUCT%%List struct {
	Page     int ` + "`" + `form:"page" binding:"required,gte=1"` + "`" + `      // 页码, 从1开始
	PageSize int ` + "`" + `form:"page_size" binding:"required,gte=1"` + "`" + ` // 每页数量, 要求大于0
	%%COL-LIST-FOR-FILTER%% %%COL-LIST-FOR-ORDER%%}

//	@Id			Get%%TABLE-NAME-STRUCT%%List
//	@Tags		%%TABLE-COMMENT%%
//	@Summary	获取%%TABLE-COMMENT%%列表
//	@Description
//	@Accept		json
//	@Produce	json
//	@Param		Authorization	header		string	true	"Bearer <jwt-token>"
//	@Param		page			query		int		true	"页码, 从1开始"
//	@Param		page_size		query		int		true	"每页数量, 要求大于0"%%COL-LIST-FOR-FILTER-DOC%%%%COL-LIST-FOR-ORDER-DOC%%
//	@Success	200				{object}	RspData{data=%%TABLE-NAME-STRUCT%%List}
//	@Failure	400				{object}	RspBase
//	@Router		/api/v1/%%TABLE-NAME-URI%% [get]
func (h *Handler) Get%%TABLE-NAME-STRUCT%%List(c *gin.Context) {
	var req ReqGet%%TABLE-NAME-STRUCT%%List
	err := shouldBind(c, &req)
	if err != nil {
		ResponseFail(c, err)
		return
	}
	ctx := c.Request.Context()
	log := contexts.GetLogger(ctx).
		WithField("req", jgstr.JsonEncode(req))
	%%PREPARE-ASSIGN-FILTER-TO-OPTION%% opt := &biz.%%TABLE-NAME-STRUCT%%ListOption{
		Pagination: &option.PaginationOption{
			Page:     req.Page,
			PageSize: req.PageSize,
		},%%ASSIGN-ORDER-TO-OPTION%% %%ASSIGN-FILTER-TO-OPTION%%
	}
	%%TABLE-NAME-VAR%%BOs, total, err := h.BizService.Get%%TABLE-NAME-STRUCT%%List(ctx, opt)
	if err != nil {
		log.WithError(err).Error("BizService.Get%%TABLE-NAME-STRUCT%%List failed")
		ResponseFail(c, err)
		return
	}
	list, err := h.To%%TABLE-NAME-STRUCT%%ListInfo(ctx, %%TABLE-NAME-VAR%%BOs)
	if err != nil {
		log.WithError(err).Error("convert %%TABLE-NAME-STRUCT%%ListBO to %%TABLE-NAME-STRUCT%%ListInfo failed")
		ResponseFail(c, err)
		return
	}
	ResponseSuccess(c, %%TABLE-NAME-STRUCT%%List{
		List:  list,
		Total: total,
	})
}
`

var TplHTTPHandlerUpdate string = `type ReqUpdate%%TABLE-NAME-STRUCT%% struct { %%COL-LIST-FOR-UPDATE%%
}

//	@Id			Update%%TABLE-NAME-STRUCT%%
//	@Tags		%%TABLE-COMMENT%%
//	@Summary	更新%%TABLE-COMMENT%%
//	@Description
//	@Accept		json
//	@Produce	json
//	@Param		Authorization	header		string			true	"Bearer <jwt-token>"
//	@Param		id				path		int				true	"%%TABLE-COMMENT%%ID"
//	@Param		body			body		ReqUpdate%%TABLE-NAME-STRUCT%%	true	"请求体"
//	@Success	200				{object}	RspBase
//	@Failure	400				{object}	RspBase
//	@Router		/api/v1/%%TABLE-NAME-URI%%/{id} [put]
func (h *Handler) Update%%TABLE-NAME-STRUCT%%(c *gin.Context) {
	id := jgstr.UintVal(c.Param("id"))
	var req ReqUpdate%%TABLE-NAME-STRUCT%%
	err := shouldBind(c, &req)
	if err != nil {
		ResponseFail(c, err)
		return
	}
	ctx := c.Request.Context()
	log := contexts.GetLogger(ctx).
		WithField("id", id).
		WithField("req", jgstr.JsonEncode(req))
	%%PREPARE-ASSIGN-UPDATE-TO-SET%% setOpt := &biz.%%TABLE-NAME-STRUCT%%SetOption{ %%ASSIGN-UPDATE-TO-SET%%
	}
	err = h.BizService.Update%%TABLE-NAME-STRUCT%%ByID(ctx, id, setOpt)
	if err != nil {
		log.WithError(err).Error("BizService.Update%%TABLE-NAME-STRUCT%%ByID failed")
		ResponseFail(c, err)
		return
	}
	ResponseSuccess(c, nil)
}
`

var TplHTTPHandlerDelete string = `// @Id			Delete%%TABLE-NAME-STRUCT%%
// @Tags		%%TABLE-COMMENT%%
// @Summary	    删除%%TABLE-COMMENT%%
// @Description
// @Accept		json
// @Produce	json
// @Param		Authorization	header		string	true	"Bearer <jwt-token>"
// @Param		id				path		int		true	"%%TABLE-COMMENT%%ID"
// @Success	200				{object}	RspBase
// @Failure	400				{object}	RspBase
// @Router		/api/v1/%%TABLE-NAME-URI%%/{id} [delete]
func (h *Handler) Delete%%TABLE-NAME-STRUCT%%(c *gin.Context) {
	id := jgstr.UintVal(c.Param("id"))
	ctx := c.Request.Context()
	log := contexts.GetLogger(ctx).WithField("id", id)
	err := h.BizService.Delete%%TABLE-NAME-STRUCT%%ByID(ctx, id)
	if err != nil {
		log.WithError(err).Error("BizService.Delete%%TABLE-NAME-STRUCT%%ByID failed")
		ResponseFail(c, err)
		return
	}
	ResponseSuccess(c, nil)
}
`

var TplHTTPDataTable string = helper.PH_META_EGG_HEADER + `
package handler

import (
	"context"
	"time"
	"%%GO-MODULE%%/internal/biz"
	jgstr "github.com/Jinglever/go-string"
	"github.com/gin-gonic/gin"
	log "%%GO-MODULE%%/pkg/log"
	"%%GO-MODULE%%/internal/common/contexts"
	"%%GO-MODULE%%/internal/common/constraint"
	"%%GO-MODULE%%/internal/repo/option"
	"%%GO-MODULE%%/internal/common/cerror"
)

// %%TABLE-COMMENT%%详情
type %%TABLE-NAME-STRUCT%%Detail struct {%%COL-LIST-IN-VO%%
}

func (h *Handler) To%%TABLE-NAME-STRUCT%%Detail(ctx context.Context, bo *biz.%%TABLE-NAME-STRUCT%%BO) (*%%TABLE-NAME-STRUCT%%Detail, error) {
	%%PREPARE-ASSIGN-BO-TO-VO%% return &%%TABLE-NAME-STRUCT%%Detail{%%ASSIGN-BO-TO-VO%%
	}, nil
}

%%TPL-HTTP-HANDLER-CREATE%%

//	@Id			Get%%TABLE-NAME-STRUCT%%Detail
//	@Tags		%%TABLE-COMMENT%%
//	@Summary	获取%%TABLE-COMMENT%%详情
//	@Description
//	@Accept		json
//	@Produce	json
//	@Param		Authorization	header		string	true	"Bearer <jwt-token>"
//	@Param		id				path		int		true	"%%TABLE-COMMENT%%ID"
//	@Success	200				{object}	RspData{data=%%TABLE-NAME-STRUCT%%Detail}
//	@Failure	400				{object}	RspBase
//	@Router		/api/v1/%%TABLE-NAME-URI%%/{id} [get]
func (h *Handler) Get%%TABLE-NAME-STRUCT%%Detail(c *gin.Context) {
	id := jgstr.UintVal(c.Param("id"))
	ctx := c.Request.Context()
	log := contexts.GetLogger(ctx).WithField("id", id)
	%%TABLE-NAME-VAR%%BO, err := h.BizService.Get%%TABLE-NAME-STRUCT%%ByID(ctx, id)
	if err != nil {
		log.WithError(err).Error("BizService.Get%%TABLE-NAME-STRUCT%%ByID failed")
		ResponseFail(c, err)
		return
	}
	d, err := h.To%%TABLE-NAME-STRUCT%%Detail(ctx, %%TABLE-NAME-VAR%%BO)
	if err != nil {
		log.WithError(err).Error("convert %%TABLE-NAME-STRUCT%%BO to %%TABLE-NAME-STRUCT%%Detail failed")
		ResponseFail(c, err)
		return
	}
	ResponseSuccess(c, d)
}

%%TPL-HTTP-HANDLER-GET-LIST%%
%%TPL-HTTP-HANDLER-UPDATE%%
%%TPL-HTTP-HANDLER-DELETE%%
`
