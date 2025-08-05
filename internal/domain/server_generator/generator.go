package svcgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"meta-egg/internal/domain/helper"
	"meta-egg/internal/domain/server_generator/template"
	"meta-egg/internal/model"

	jgstr "github.com/Jinglever/go-string"
	log "github.com/sirupsen/logrus"
)

// relativeDir2NeedConfirm: key是相对于工程根目录的路径，值是布尔值，代表是否需要确认；注意，不会递归查看子目录
func Generate(codeDir string, project *model.Project) (relativeDir2NeedConfirm map[string]bool, err error) {
	var (
		hasGRPC bool
		hasHTTP bool
	)
	if project.ServerType == model.ServerType_GRPC {
		hasGRPC = true
	} else if project.ServerType == model.ServerType_HTTP {
		hasHTTP = true
	} else {
		hasGRPC = true
		hasHTTP = true
	}
	relativeDir2NeedConfirm = make(map[string]bool)
	if hasGRPC {
		relativeDir2NeedConfirm[filepath.Join("internal", "server", "grpc")] = true
	}
	if hasHTTP {
		relativeDir2NeedConfirm[filepath.Join("internal", "server", "http")] = true
	}
	relativeDir2NeedConfirm[filepath.Join("internal", "server", "monitor")] = true
	// 创建目录
	for dir := range relativeDir2NeedConfirm {
		path := filepath.Join(codeDir, dir)
		if err = os.MkdirAll(path, 0755); err != nil {
			log.Errorf("failed to mkdir %s: %v", dir, err)
			return
		}
	}

	var path string

	if hasGRPC {
		// internal/server/grpc/middleware.go
		path = filepath.Join(codeDir, "internal", "server", "grpc", "middleware.go")
		err = generateGoFile(path, template.TplInternalServerGRPCMiddleware, project, helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate internal/server/grpc/middleware.go failed: %v", err)
			return
		}

		// internal/server/grpc/server.go
		path = filepath.Join(codeDir, "internal", "server", "grpc", "server.go")
		err = generateGoFile(path, template.TplInternalServerGRPCServer, project, helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate internal/server/grpc/server.go failed: %v", err)
			return
		}
	}

	if hasHTTP {
		// internal/server/http/middleware.go
		path = filepath.Join(codeDir, "internal", "server", "http", "middleware.go")
		err = generateGoFile(path, template.TplInternalServerHTTPMiddleware, project, helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate internal/server/http/middleware.go failed: %v", err)
			return
		}

		// internal/server/http/router.go
		path = filepath.Join(codeDir, "internal", "server", "http", "router.go")
		err = generateGoFile(path, template.TplInternalServerHTTPRouter, project, helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate internal/server/http/router.go failed: %v", err)
			return
		}

		// internal/server/http/server.go
		path = filepath.Join(codeDir, "internal", "server", "http", "server.go")
		err = generateGoFile(path, template.TplInternalServerHTTPServer, project, helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate internal/server/http/server.go failed: %v", err)
			return
		}
	}

	// // internal/server/monitor/router.go
	// path = filepath.Join(codeDir, "internal", "server", "monitor", "router.go")
	// err = generateGoFile(path, template.TplInternalServerMonitorRouter, project, helper.AddHeaderCanEdit)
	// if err != nil {
	// 	log.Errorf("generate internal/server/monitor/router.go failed: %v", err)
	// 	return
	// }

	// internal/server/monitor/server.go
	path = filepath.Join(codeDir, "internal", "server", "monitor", "server.go")
	err = generateGoFile(path, template.TplInternalServerMonitorServer, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate internal/server/monitor/server.go failed: %v", err)
		return
	}

	return
}

func generateGoFile(path string, tpl string, project *model.Project, addHeader func(s string) string) error {
	f, err := os.Create(path)
	if err != nil {
		log.Errorf("create file %s failed: %v", path, err)
		return err
	}
	defer f.Close()

	// template
	code := tpl
	if addHeader != nil {
		code = addHeader(code)
	}
	replaceTpl(&code, project)

	// go format
	formatted, err := jgstr.FormatGo([]byte(code))
	if err != nil {
		log.Errorf("format source failed: %v\n%s", err, code)
		return err
	}
	f.Write(formatted)
	return nil
}

func replaceTpl(code *string, project *model.Project) {
	genHTTPRouteMapping(code, project)
	genGRPCMiddleware(code, project)
	genHTTPMiddleware(code, project)

	if project.NoAuth {
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_SERVER_CONFIG_ACCESS_TOKEN, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_SERVER_CONFIG_ACCESS_TOKEN, "")
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_SERVER_CONFIG_ACCESS_TOKEN, template.TplGRPCServerConfigAccessToken)
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_SERVER_CONFIG_ACCESS_TOKEN, template.TplHTTPServerConfigAccessToken)
	}
	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, project.GoModule)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_DIR, helper.GetDirName(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_STRUCT, helper.GetStructName(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME, project.Name)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_PKG, helper.GetPkgName(project.Name))

	if project.Database == nil || len(project.Database.Tables) == 0 {
		*code = strings.ReplaceAll(*code, template.PH_COMMENT_REPO, "//")
		if len(project.Domain.Usecases) == 0 {
			*code = strings.ReplaceAll(*code, template.PH_COMMENT_DOMAIN, "//")
		} else {
			*code = strings.ReplaceAll(*code, template.PH_COMMENT_DOMAIN, "")
		}
	} else {
		*code = strings.ReplaceAll(*code, template.PH_COMMENT_REPO, "")
		if len(project.Domain.Usecases) == 0 {
			*code = strings.ReplaceAll(*code, template.PH_COMMENT_DOMAIN, "//")
		} else {
			*code = strings.ReplaceAll(*code, template.PH_COMMENT_DOMAIN, "")
		}
	}
}

func replaceTplForTable(code *string, table *model.Table) {
	*code = strings.ReplaceAll(*code, template.PH_TABLE_COMMENT, table.Comment)
	*code = strings.ReplaceAll(*code, template.PH_TABLE_NAME_URI, helper.GetURIName(table.Name))
	*code = strings.ReplaceAll(*code, template.PH_TABLE_NAME_STRUCT, helper.GetStructName(table.Name))
}

func genHTTPRouteMapping(code *string, project *model.Project) {
	if project.ServerType != model.ServerType_HTTP &&
		project.ServerType != model.ServerType_ALL {
		return
	}
	var buf strings.Builder

	// in case zero handler
	zeroHandler := true
	buf.WriteString(template.TplHTTPRouteZeroHandler)

	if project.Database != nil {
		for _, table := range project.Database.Tables {
			if !table.HasHandler {
				continue
			}
			if zeroHandler {
				zeroHandler = false
				buf.Reset()
				buf.WriteString(template.TplHTTPRouteNewHandler)
			}
			if table.Type == model.TableType_DATA {
				str := template.TplHTTPRouteMappingForDataTable
				replaceTplForTable(&str, table)
				// 添加RL表路由
				genRLHTTPRoutes(&str, table, project)
				buf.WriteString(str)
			} else if table.Type == model.TableType_META {
				str := template.TplHTTPRouteMappingForMetaTable
				replaceTplForTable(&str, table)
				buf.WriteString(str)
			} else if table.Type == model.TableType_BR {
				// 为BR表生成路由
				str := genBRHTTPRoutes(table, project)
				buf.WriteString(str)
			}
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_HTTP_ROUTE_MAPPING, buf.String())
	if zeroHandler {
		*code = strings.ReplaceAll(*code, template.PH_IMPORT_HDL_COMMENT, "//")
	} else {
		*code = strings.ReplaceAll(*code, template.PH_IMPORT_HDL_COMMENT, "")
	}

	if project.NoAuth {
		*code = strings.ReplaceAll(*code, template.PH_HTTP_ROUTE_USE_AUTH_HANDLER, "")
	} else {
		*code = strings.ReplaceAll(*code, template.PH_HTTP_ROUTE_USE_AUTH_HANDLER, template.TplHTTPRouterUseAuthHandler)
	}
}

func genGRPCMiddleware(code *string, project *model.Project) {
	if project.ServerType != model.ServerType_GRPC &&
		project.ServerType != model.ServerType_ALL {
		return
	}
	if project.NoAuth {
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_GRPC_AUTH_INTERCEPTOR, template.TplFuncGRPCExtractME)
		*code = strings.ReplaceAll(*code, template.PH_TPL_CALL_FUNC_GRPC_AUTH_INTERCEPTOR, template.TplCallFuncGRPCExtractME)
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_GRPC_AUTH_INTERCEPTOR, template.TplFuncGRPCAuthInterceptor)
		*code = strings.ReplaceAll(*code, template.PH_TPL_CALL_FUNC_GRPC_AUTH_INTERCEPTOR, template.TplCallFuncGRPCAuthInterceptor)
	}
}

func genHTTPMiddleware(code *string, project *model.Project) {
	if project.ServerType != model.ServerType_HTTP &&
		project.ServerType != model.ServerType_ALL {
		return
	}
	if project.NoAuth {
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_HTTP_AUTH_HANDLER, "")
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_HTTP_AUTH_HANDLER, template.TplFuncHTTPAuthHandler)
	}
}

func genRLHTTPRoutes(code *string, mainTable *model.Table, project *model.Project) {
	if mainTable.Type != model.TableType_DATA {
		*code = strings.ReplaceAll(*code, template.PH_RL_HTTP_ROUTES, "")
		return
	}

	// 构建表名到表的映射
	tableNameToTable := make(map[string]*model.Table)
	for _, table := range project.Database.Tables {
		tableNameToTable[table.Name] = table
	}

	// 获取该主表的所有RL表
	var rlTables []*model.Table
	for _, table := range project.Database.Tables {
		if table.Type == model.TableType_RL {
			// 查找指向主表的外键
			identifiedMainTable := helper.IdentifyRLMainTable(table, tableNameToTable)
			if identifiedMainTable != nil && identifiedMainTable.Name == mainTable.Name {
				rlTables = append(rlTables, table)
			}
		}
	}

	// 生成RL表路由
	var buf strings.Builder
	for _, rlTable := range rlTables {
		str := template.TplHTTPRLRoutes
		replaceTplForRLRoute(&str, mainTable, rlTable)
		buf.WriteString(str)
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_HTTP_ROUTES, buf.String())
}

// replaceTplForRLRoute 为RL表路由替换占位符
func replaceTplForRLRoute(code *string, mainTable *model.Table, rlTable *model.Table) {
	*code = strings.ReplaceAll(*code, template.PH_MAIN_TABLE_COMMENT, mainTable.Comment)
	*code = strings.ReplaceAll(*code, template.PH_MAIN_TABLE_NAME_URI, helper.GetURIName(mainTable.Name))
	*code = strings.ReplaceAll(*code, template.PH_MAIN_TABLE_NAME, strings.ToLower(mainTable.Name))
	*code = strings.ReplaceAll(*code, template.PH_RL_TABLE_NAME_STRUCT, helper.GetStructName(rlTable.Name))
	*code = strings.ReplaceAll(*code, template.PH_RL_TABLE_COMMENT, rlTable.Comment)
	*code = strings.ReplaceAll(*code, template.PH_RL_TABLE_NAME_URI, helper.GetURIName(rlTable.Name))
	*code = strings.ReplaceAll(*code, template.PH_RL_TABLE_NAME, strings.ToLower(rlTable.Name))
}

// genBRHTTPRoutes 为BR表生成HTTP路由
func genBRHTTPRoutes(brTable *model.Table, project *model.Project) string {
	// 构建表名到表的映射
	tableNameToTable := make(map[string]*model.Table)
	for _, table := range project.Database.Tables {
		tableNameToTable[table.Name] = table
	}

	// 识别BR表的两个关联表
	brRelatedTables := helper.IdentifyBRRelatedTables(brTable, tableNameToTable)
	if brRelatedTables == nil {
		return ""
	}

	var buf strings.Builder

	// 生成注释
	buf.WriteString(fmt.Sprintf("\n\t\t// %s", brTable.Comment))

	// 生成双向查询路由
	// Table1 -> Table2 的查询路由
	buf.WriteString(fmt.Sprintf(`
		apiGroup.GET("/%s/{%s_id}/%s", handler.Get%sListBy%sID)`,
		helper.GetURIName(brRelatedTables.Table1.Name),     // /users
		helper.GetVarName(brRelatedTables.Table1.Name),     // user_id
		helper.GetURIName(brRelatedTables.Table2.Name),     // /tags
		helper.GetStructName(brRelatedTables.Table2.Name),  // Tag
		helper.GetStructName(brRelatedTables.Table1.Name))) // User

	// Table2 -> Table1 的查询路由
	buf.WriteString(fmt.Sprintf(`
		apiGroup.GET("/%s/{%s_id}/%s", handler.Get%sListBy%sID)`,
		helper.GetURIName(brRelatedTables.Table2.Name),     // /tags
		helper.GetVarName(brRelatedTables.Table2.Name),     // tag_id
		helper.GetURIName(brRelatedTables.Table1.Name),     // /users
		helper.GetStructName(brRelatedTables.Table1.Name),  // User
		helper.GetStructName(brRelatedTables.Table2.Name))) // Tag

	// 生成批量绑定路由
	// 给Table1批量分配Table2
	table2PluralName := helper.GetPluralName(brRelatedTables.Table2.Name)
	buf.WriteString(fmt.Sprintf(`
		apiGroup.POST("/%s/{%s_id}/bind-%s", handler.Bind%sTo%s)`,
		helper.GetURIName(brRelatedTables.Table1.Name),     // /users
		helper.GetVarName(brRelatedTables.Table1.Name),     // user_id
		helper.GetURIName(table2PluralName),                // /tags
		helper.GetStructName(table2PluralName),             // Tags
		helper.GetStructName(brRelatedTables.Table1.Name))) // User

	// 给Table2批量分配Table1
	table1PluralName := helper.GetPluralName(brRelatedTables.Table1.Name)
	buf.WriteString(fmt.Sprintf(`
		apiGroup.POST("/%s/{%s_id}/bind-%s", handler.Bind%sTo%s)`,
		helper.GetURIName(brRelatedTables.Table2.Name),     // /tags
		helper.GetVarName(brRelatedTables.Table2.Name),     // tag_id
		helper.GetURIName(table1PluralName),                // /users
		helper.GetStructName(table1PluralName),             // Users
		helper.GetStructName(brRelatedTables.Table2.Name))) // Tag

	// 生成批量解绑路由
	// 从Table1解绑Table2
	buf.WriteString(fmt.Sprintf(`
		apiGroup.POST("/%s/{%s_id}/unbind-%s", handler.Unbind%sFrom%s)`,
		helper.GetURIName(brRelatedTables.Table1.Name),     // /users
		helper.GetVarName(brRelatedTables.Table1.Name),     // user_id
		helper.GetURIName(table2PluralName),                // /tags
		helper.GetStructName(table2PluralName),             // Tags
		helper.GetStructName(brRelatedTables.Table1.Name))) // User

	// 从Table2解绑Table1
	buf.WriteString(fmt.Sprintf(`
		apiGroup.POST("/%s/{%s_id}/unbind-%s", handler.Unbind%sFrom%s)`,
		helper.GetURIName(brRelatedTables.Table2.Name),     // /tags
		helper.GetVarName(brRelatedTables.Table2.Name),     // tag_id
		helper.GetURIName(table1PluralName),                // /users
		helper.GetStructName(table1PluralName),             // Users
		helper.GetStructName(brRelatedTables.Table2.Name))) // Tag

	return buf.String()
}
