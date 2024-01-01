package template

import "meta-egg/internal/domain/helper"

var TplHTTPRouteNewHandler = `handler := hdl.WireHandler(s.Resource)
`

var TplHTTPRouteMappingForDataTable = `
	// %%TABLE-COMMENT%%
	apiGroup.POST("/%%TABLE-NAME-URI%%", handler.Create%%TABLE-NAME-STRUCT%%)
	apiGroup.GET("/%%TABLE-NAME-URI%%/:id", handler.Get%%TABLE-NAME-STRUCT%%Detail)
	apiGroup.GET("/%%TABLE-NAME-URI%%", handler.Get%%TABLE-NAME-STRUCT%%List)
	apiGroup.PUT("/%%TABLE-NAME-URI%%/:id", handler.Update%%TABLE-NAME-STRUCT%%)
	apiGroup.DELETE("/%%TABLE-NAME-URI%%/:id", handler.Delete%%TABLE-NAME-STRUCT%%)
`

var TplHTTPRouteMappingForMetaTable = `
	// %%TABLE-COMMENT%%
	apiGroup.GET("/%%TABLE-NAME-URI%%/:id", handler.Get%%TABLE-NAME-STRUCT%%Detail)
	apiGroup.GET("/%%TABLE-NAME-URI%%", handler.Get%%TABLE-NAME-STRUCT%%List)
`

var TplHTTPRouteZeroHandler = `// handler := hdl.WireHandler(s.Resource)

// 用户
// TODO: add your router mapping here
// Such as:
//   apiGroup.GET("/users/:id", handler.GetUserDetail)	
`

var TplInternalServerHTTPRouter string = helper.PH_META_EGG_HEADER + `
package server

import (
	"github.com/gin-gonic/gin"
	%%IMPORT-HDL-COMMENT%%hdl "%%GO-MODULE%%/internal/handler/http"
)

// annotation for swagger
// @title	%%PROJECT-NAME%%
// @version xxx
func (s *Server) initRouter() {
	router := gin.New()
	router.Use(Logger(), gin.Recovery())
	router.Use(errorHandler(s.Cfg))
	apiGroup := router.Group("/api/v1")
	%%HTTP-ROUTE-USE-AUTH-HANDLER%%

	%%HTTP-ROUTE-MAPPING%%

	s.Router = router
}
`

var TplHTTPRouterUseAuthHandler string = `// skip authHandler for some path
skipFullPath := map[string]struct{}{}
apiGroup.Use(authHandler(s.Resource.AccessToken, s.Cfg, skipFullPath))`
