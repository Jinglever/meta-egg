package cmdgen

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"meta-egg/internal/domain/cmd_generator/template"
	"meta-egg/internal/domain/helper"
	"meta-egg/internal/model"

	jgstr "github.com/Jinglever/go-string"
)

// relativeDir2NeedConfirm: key是相对于工程根目录的路径，值是布尔值，代表是否需要确认；注意，不会递归查看子目录
func Generate(codeDir string, project *model.Project) (relativeDir2NeedConfirm map[string]bool, err error) {
	relativeDir2NeedConfirm = map[string]bool{
		filepath.Join("cmd", project.Name): true,
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

	// main.go
	path = filepath.Join(codeDir, "cmd", project.Name, "main.go")
	err = generateGoFile(path, template.TplMain, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate main.go failed: %v", err)
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
	if project.ServerType == model.ServerType_HTTP {
		*code = strings.ReplaceAll(*code, template.PH_MAIN_RUN_GRPC_SERVER, "")
		*code = strings.ReplaceAll(*code, template.PH_MAIN_CANCEL_GRPC_SERVER, "")
	} else if project.ServerType == model.ServerType_GRPC {
		*code = strings.ReplaceAll(*code, template.PH_MAIN_RUN_HTTP_SERVER, "")
		*code = strings.ReplaceAll(*code, template.PH_MAIN_CANCEL_HTTP_SERVER, "")
	}
	*code = strings.ReplaceAll(*code, template.PH_MAIN_RUN_HTTP_SERVER, template.TplMainRunHttpServer)
	*code = strings.ReplaceAll(*code, template.PH_MAIN_CANCEL_HTTP_SERVER, template.TplMainCancelHttpServer)
	*code = strings.ReplaceAll(*code, template.PH_MAIN_RUN_GRPC_SERVER, template.TplMainRunGrpcServer)
	*code = strings.ReplaceAll(*code, template.PH_MAIN_CANCEL_GRPC_SERVER, template.TplMainCancelGrpcServer)

	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, project.GoModule)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME, project.Name)
}
