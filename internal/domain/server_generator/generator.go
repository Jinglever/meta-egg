package svcgen

import (
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
				buf.WriteString(str)
			} else if table.Type == model.TableType_META {
				str := template.TplHTTPRouteMappingForMetaTable
				replaceTplForTable(&str, table)
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
