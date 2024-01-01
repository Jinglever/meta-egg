package testgen

import (
	"os"
	"path/filepath"
	"strings"

	"meta-egg/internal/domain/helper"
	"meta-egg/internal/domain/test_generator/template"
	"meta-egg/internal/model"

	log "github.com/sirupsen/logrus"

	jgstr "github.com/Jinglever/go-string"
)

// relativeDir2NeedConfirm: key是相对于工程根目录的路径，值是布尔值，代表是否需要确认；注意，不会递归查看子目录
func Generate(codeDir string, project *model.Project) (relativeDir2NeedConfirm map[string]bool, err error) {
	relativeDir2NeedConfirm = map[string]bool{
		filepath.Join("test", "integrate"): true,
		filepath.Join("test", "unit"):      true,
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

	// test/integrate/helper.go
	path = filepath.Join(codeDir, "test", "integrate", "helper.go")
	err = generateGoFile(path, template.TplIntegrateHelper, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate test/integrate/helper.go failed: %v", err)
		return
	}

	// test/integrate/helper_test.go
	path = filepath.Join(codeDir, "test", "integrate", "helper_test.go")
	err = generateGoFile(path, template.TplIntegrateHelperTest, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate test/integrate/helper_test.go failed: %v", err)
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
	_, _ = f.Write(formatted)
	return nil
}

func replaceTpl(code *string, project *model.Project) {
	replaceResourceTpl(code, project)

	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, project.GoModule)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME, project.Name)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_DIR, helper.GetDirName(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_PKG, helper.GetPkgName(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_STRUCT, helper.GetStructName(project.Name))
}

func replaceResourceTpl(code *string, project *model.Project) {
	/*
		if project.Database == nil {
			*code = strings.ReplaceAll(*code, template.PH_TPL_RESOURCE_STRUCT_DB, "")
			*code = strings.ReplaceAll(*code, template.PH_TPL_RESOURCE_CONFIG_STRUCT_DB, "")
			*code = strings.ReplaceAll(*code, template.PH_TPL_RESOURCE_DB, "")
		} else {
			*code = strings.ReplaceAll(*code, template.PH_TPL_RESOURCE_STRUCT_DB, template.TplResourceStructDB)
			*code = strings.ReplaceAll(*code, template.PH_TPL_RESOURCE_CONFIG_STRUCT_DB, template.TplResourceConfigStructDB)
			*code = strings.ReplaceAll(*code, template.PH_TPL_RESOURCE_DB, template.TplResourceDB)
		}
	*/
}
