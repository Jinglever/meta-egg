package cfggen

import (
	"os"
	"path/filepath"
	"strings"

	"meta-egg/internal/domain/config_generator/template"
	"meta-egg/internal/domain/helper"
	"meta-egg/internal/model"

	log "github.com/sirupsen/logrus"

	jgstr "github.com/Jinglever/go-string"
)

// relativeDir2NeedConfirm: key是相对于工程根目录的路径，值是布尔值，代表是否需要确认；注意，不会递归查看子目录
func Generate(codeDir string, project *model.Project) (relativeDir2NeedConfirm map[string]bool, err error) {
	relativeDir2NeedConfirm = map[string]bool{
		"configs":                           true,
		filepath.Join("internal", "config"): true,
	}
	// 创建目录
	for dir := range relativeDir2NeedConfirm {
		path := filepath.Join(codeDir, dir)
		if err = os.MkdirAll(path, 0755); err != nil {
			log.Errorf("failed to mkdir %s: %v", dir, err)
			return
		}
	}

	var path string

	// conf.yml
	path = filepath.Join(codeDir, "configs", project.Name+".yml")
	err = generateNonGoFile(path, template.TplConfYml, project, nil)
	if err != nil {
		log.Errorf("generate configs/%v.yml failed: %v", project.Name, err)
		return
	}

	// internal/config/config.go
	path = filepath.Join(codeDir, "internal", "config", "config.go")
	err = generateGoFile(path, template.TplInternalConfigConfig, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate internal/config/config.go failed: %v", err)
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
	if project.Database == nil {
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_RESOURCE_DB, "")
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_RESOURCE_DB, template.TplConfYmlResourceDB)
	}
	if project.NoAuth {
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_RESOURCE_ACCESS_TOKEN, "")
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_RESOURCE_ACCESS_TOKEN, template.TplConfYmlResourceAccessToken)
	}

	if project.ServerType == model.ServerType_GRPC {
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_HTTP_SERVER, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_GRPC_SERVER, template.TplConfYmlGrpcServer)

		*code = strings.ReplaceAll(*code, template.PH_TPL_CONFIG_STRUCT_HTTP_SERVER, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONFIG_FUNC_GET_HTTP_SERVER, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONFIG_STRUCT_GRPC_SERVER, template.TplConfigStructGrpcServer)
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONFIG_FUNC_GET_GRPC_SERVER, template.TplConfigFuncGetGrpcServer)
	} else if project.ServerType == model.ServerType_HTTP {
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_HTTP_SERVER, template.TplConfYmlHttpServer)
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_GRPC_SERVER, "")

		*code = strings.ReplaceAll(*code, template.PH_TPL_CONFIG_STRUCT_HTTP_SERVER, template.TplConfigStructHttpServer)
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONFIG_FUNC_GET_HTTP_SERVER, template.TplConfigFuncGetHttpServer)
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONFIG_STRUCT_GRPC_SERVER, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONFIG_FUNC_GET_GRPC_SERVER, "")
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_HTTP_SERVER, template.TplConfYmlHttpServer)
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_GRPC_SERVER, template.TplConfYmlGrpcServer)

		*code = strings.ReplaceAll(*code, template.PH_TPL_CONFIG_STRUCT_HTTP_SERVER, template.TplConfigStructHttpServer)
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONFIG_FUNC_GET_HTTP_SERVER, template.TplConfigFuncGetHttpServer)
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONFIG_STRUCT_GRPC_SERVER, template.TplConfigStructGrpcServer)
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONFIG_FUNC_GET_GRPC_SERVER, template.TplConfigFuncGetGrpcServer)
	}

	if project.NoAuth {
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_GRPC_SERVER_ACCESS_TOKEN, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_HTTP_SERVER_ACCESS_TOKEN, "")
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_GRPC_SERVER_ACCESS_TOKEN, template.TplConfYmlGRPCServerAccessToken)
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONF_YML_HTTP_SERVER_ACCESS_TOKEN, template.TplConfYmlHTTPServerAccessToken)
	}

	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, project.GoModule)
	*code = strings.ReplaceAll(*code, template.PH_ENV_PREFIX, helper.GetEnvPrefix(project.Name))

	if project.Database != nil {
		if project.Database.Type == model.DBType_MYSQL || project.Database.Type == model.DBType_TIDB {
			*code = strings.ReplaceAll(*code, template.PH_TPL_DB_TYPE, template.TplDbTypeMysql)
			*code = strings.ReplaceAll(*code, template.PH_TPL_DB_DSN, template.TplDbDsnMysql)
		} else if project.Database.Type == model.DBType_PG {
			*code = strings.ReplaceAll(*code, template.PH_TPL_DB_TYPE, template.TplDbTypePostgres)
			*code = strings.ReplaceAll(*code, template.PH_TPL_DB_DSN, template.TplDbDsnPostgres)
		}
	}
}

func generateNonGoFile(path string, tpl string, project *model.Project, addHeader func(s string) string) error {
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

	f.Write([]byte(code))
	return nil
}
