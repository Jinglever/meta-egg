package template

import "meta-egg/internal/domain/helper"

var TplGRPCHandlerCreate string = `// 创建%%TABLE-COMMENT%%
func (h *Handler) Create%%TABLE-NAME-STRUCT%%(ctx context.Context,
	req *api.Create%%TABLE-NAME-STRUCT%%Request,
) (*api.%%TABLE-NAME-STRUCT%%Detail, error) {
	log := contexts.GetLogger(ctx).
		WithField("req", jgstr.JsonEncode(req))
	err := req.ValidateAll()
	if err != nil {
		log.WithError(err).Error("req.ValidateAll failed")
		return nil, cerror.InvalidArgument(err.Error())
	}
	%%PREPARE-ASSIGN-CREATE-TO-BO-GRPC%% %%TABLE-NAME-VAR%%BO := &biz.%%TABLE-NAME-STRUCT%%BO{%%ASSIGN-CREATE-TO-BO-GRPC%%
	}
	err = h.BizService.Create%%TABLE-NAME-STRUCT%%(ctx, %%TABLE-NAME-VAR%%BO)
	if err != nil {
		log.WithError(err).Error("BizService.Create%%TABLE-NAME-STRUCT%% failed")
		return nil, err
	}
	d, err := h.To%%TABLE-NAME-STRUCT%%Detail(ctx, %%TABLE-NAME-VAR%%BO)
	if err != nil {
		log.WithError(err).Error("convert %%TABLE-NAME-STRUCT%%BO to %%TABLE-NAME-STRUCT%%Detail failed")
		return nil, err
	}
	return d, nil
}
`

var TplGRPCHandlerGetList string = `func (h *Handler) To%%TABLE-NAME-STRUCT%%ListInfo(ctx context.Context,
	objs []*biz.%%TABLE-NAME-STRUCT%%ListBO,
) ([]*api.%%TABLE-NAME-STRUCT%%ListInfo, error) {
	list := make([]*api.%%TABLE-NAME-STRUCT%%ListInfo, 0, len(objs))
	for i := range objs {
		%%PREPARE-ASSIGN-BO-FOR-LIST%% list = append(list, &api.%%TABLE-NAME-STRUCT%%ListInfo{%%ASSIGN-BO-FOR-LIST%%
		})
	}
	return list, nil
}

// 获取%%TABLE-COMMENT%%列表
func (h *Handler) Get%%TABLE-NAME-STRUCT%%List(ctx context.Context,
	req *api.Get%%TABLE-NAME-STRUCT%%ListRequest,
) (*api.Get%%TABLE-NAME-STRUCT%%ListResponse, error) {
	log := contexts.GetLogger(ctx).
		WithField("req", jgstr.JsonEncode(req))
	err := req.ValidateAll()
	if err != nil {
		log.WithError(err).Error("req.ValidateAll failed")
		return nil, cerror.InvalidArgument(err.Error())
	}
	
	%%PREPARE-ASSIGN-FILTER-TO-OPTION-GRPC%% opt := &biz.%%TABLE-NAME-STRUCT%%ListOption{%%ASSIGN-ORDER-TO-OPTION%% %%ASSIGN-FILTER-TO-OPTION-GRPC%%
	}
	if req.Pagination != nil {
		opt.Pagination = &option.PaginationOption{
			Page:     int(req.Pagination.Page),
			PageSize: int(req.Pagination.PageSize),
		}
	}
	%%TABLE-NAME-VAR%%BOs, total, err := h.BizService.Get%%TABLE-NAME-STRUCT%%List(ctx, opt)
	if err != nil {
		log.WithError(err).Error("BizService.Get%%TABLE-NAME-STRUCT%%List failed")
		return nil, err
	}
	list, err := h.To%%TABLE-NAME-STRUCT%%ListInfo(ctx, %%TABLE-NAME-VAR%%BOs)
	if err != nil {
		log.WithError(err).Error("convert %%TABLE-NAME-STRUCT%%ListBO to %%TABLE-NAME-STRUCT%%ListInfo failed")
		return nil, err
	}
	return &api.Get%%TABLE-NAME-STRUCT%%ListResponse{
		List:  list,
		Total: total,
	}, nil
}
`

var TplGRPCHandlerDelete string = `// 删除%%TABLE-COMMENT%%
func (h *Handler) Delete%%TABLE-NAME-STRUCT%%(ctx context.Context,
	req *api.Delete%%TABLE-NAME-STRUCT%%Request,
) (*emptypb.Empty, error) {
	log := contexts.GetLogger(ctx).
		WithField("req", jgstr.JsonEncode(req))
	err := req.ValidateAll()
	if err != nil {
		log.WithError(err).Error("req.ValidateAll failed")
		return nil, cerror.InvalidArgument(err.Error())
	}
	err = h.BizService.Delete%%TABLE-NAME-STRUCT%%ByID(ctx, req.Id)
	if err != nil {
		log.WithError(err).Error("BizService.Delete%%TABLE-NAME-STRUCT%%ByID failed")
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
`

var TplGRPCHandlerUpdate string = `// 更新%%TABLE-COMMENT%%
func (h *Handler) Update%%TABLE-NAME-STRUCT%%(ctx context.Context,
	req *api.Update%%TABLE-NAME-STRUCT%%Request,
) (*emptypb.Empty, error) {
	log := contexts.GetLogger(ctx).
		WithField("req", jgstr.JsonEncode(req))
	err := req.ValidateAll()
	if err != nil {
		log.WithError(err).Error("req.ValidateAll failed")
		return nil, cerror.InvalidArgument(err.Error())
	}	
	%%PREPARE-ASSIGN-UPDATE-TO-SET-GRPC%% setOpt := &biz.%%TABLE-NAME-STRUCT%%SetOption{ %%ASSIGN-UPDATE-TO-SET-GRPC%%
	}
	err = h.BizService.Update%%TABLE-NAME-STRUCT%%ByID(ctx, req.Id, setOpt)
	if err != nil {
		log.WithError(err).Error("BizService.Update%%TABLE-NAME-STRUCT%%ByID failed")
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
`

var TplGRPCTable string = helper.PH_META_EGG_HEADER + `
package handler

import (
	"context"
	"time"
	api "%%GO-MODULE%%/api/%%PROJECT-NAME-DIR%%"
	"%%GO-MODULE%%/internal/biz"
	jgstr "github.com/Jinglever/go-string"
	"github.com/gin-gonic/gin"
	log "%%GO-MODULE%%/pkg/log"
	"%%GO-MODULE%%/internal/common/contexts"
	"%%GO-MODULE%%/internal/common/constraint"
	"%%GO-MODULE%%/internal/repo/option"
	"%%GO-MODULE%%/internal/common/cerror"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *Handler) To%%TABLE-NAME-STRUCT%%Detail(ctx context.Context,
	bo *biz.%%TABLE-NAME-STRUCT%%BO,
) (*api.%%TABLE-NAME-STRUCT%%Detail, error) {
	%%PREPARE-ASSIGN-BO-TO-VO%% return &api.%%TABLE-NAME-STRUCT%%Detail{%%ASSIGN-BO-TO-VO-GRPC%%
	}, nil
}

%%TPL-GRPC-HANDLER-CREATE%%

// 获取%%TABLE-COMMENT%%详情
func (h *Handler) Get%%TABLE-NAME-STRUCT%%Detail(ctx context.Context,
	req *api.Get%%TABLE-NAME-STRUCT%%DetailRequest,
) (*api.%%TABLE-NAME-STRUCT%%Detail, error) {
	log := contexts.GetLogger(ctx).
		WithField("req", jgstr.JsonEncode(req))
	%%TABLE-NAME-VAR%%BO, err := h.BizService.Get%%TABLE-NAME-STRUCT%%ByID(ctx, req.Id)
	if err != nil {
		log.WithError(err).Error("BizService.Get%%TABLE-NAME-STRUCT%%ByID failed")
		return nil, err
	}
	d, err := h.To%%TABLE-NAME-STRUCT%%Detail(ctx, %%TABLE-NAME-VAR%%BO)
	if err != nil {
		log.WithError(err).Error("convert %%TABLE-NAME-STRUCT%%BO to %%TABLE-NAME-STRUCT%%Detail failed")
		return nil, err
	}
	return d, nil
}

%%TPL-GRPC-HANDLER-GET-LIST%%
%%TPL-GRPC-HANDLER-UPDATE%%
%%TPL-GRPC-HANDLER-DELETE%%
`
