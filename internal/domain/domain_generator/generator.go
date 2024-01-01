package domaingen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"meta-egg/internal/domain/domain_generator/template"
	"meta-egg/internal/domain/helper"
	"meta-egg/internal/model"

	jgstr "github.com/Jinglever/go-string"
	log "github.com/sirupsen/logrus"
)

// relativeDir2NeedConfirm: key是相对于工程根目录的路径，值是布尔值，代表是否需要确认；注意，不会递归查看子目录
func Generate(codeDir string, project *model.Project) (relativeDir2NeedConfirm map[string]bool, err error) {
	relativeDir2NeedConfirm = map[string]bool{
		filepath.Join("internal", "usecase"): true,
	}
	for _, usecase := range project.Domain.Usecases {
		relativeDir2NeedConfirm[filepath.Join("internal", "usecase", helper.GetDirName(usecase.Name))] = true
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

	// internal/usecase/base.go
	path = filepath.Join(codeDir, "internal", "usecase", "base.go")
	err = generateGoFile(path, template.TplBase, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate internal/usecase/base.go failed: %v", err)
		return
	}

	// for usecase
	for _, usecase := range project.Domain.Usecases {
		// internal/usecase/<usecase>/base.go
		path = filepath.Join(codeDir, "internal", "usecase", helper.GetDirName(usecase.Name), "base.go")
		err = generateGoFileForUsecase(path, template.TplUsecaseBase, usecase, helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate internal/usecase/%s/base.go failed: %v", helper.GetDirName(usecase.Name), err)
			return
		}
	}

	return
}

func generateGoFileForUsecase(path string, tpl string, usecase *model.Usecase, addHeader func(s string) string) error {
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
	replaceTplForUsecase(&code, usecase)

	// go format
	formatted, err := jgstr.FormatGo([]byte(code))
	if err != nil {
		log.Errorf("format source failed: %v\n%s", err, code)
		return err
	}
	_, _ = f.Write(formatted)
	return nil
}

func replaceTplForUsecase(code *string, usecase *model.Usecase) {
	genImportUsecaseList(code, usecase.Domain.Project)
	genProviderUsecaseList(code, usecase.Domain.Project)

	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, usecase.Domain.Project.GoModule)
	*code = strings.ReplaceAll(*code, template.PH_USECASE_NAME_PKG, helper.GetPkgName(usecase.Name))
	*code = strings.ReplaceAll(*code, template.PH_USECASE_NAME_STRUCT, helper.GetStructName(usecase.Name))
	*code = strings.ReplaceAll(*code, template.PH_USECASE_DESC, usecase.Desc)

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
	genImportUsecaseList(code, project)
	genProviderUsecaseList(code, project)

	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, project.GoModule)
}

func genImportUsecaseList(code *string, project *model.Project) {
	// import usecase list
	var buf strings.Builder
	for _, usecase := range project.Domain.Usecases {
		buf.WriteString(fmt.Sprintf("%s \"%s/internal/usecase/%s\"\n",
			helper.GetPkgName(usecase.Name),
			project.GoModule,
			helper.GetDirName(usecase.Name)),
		)
	}
	*code = strings.ReplaceAll(*code, template.PH_IMPORT_USECASE_LIST, buf.String())
}

func genProviderUsecaseList(code *string, project *model.Project) {
	// provider usecase list
	var buf strings.Builder
	for _, usecase := range project.Domain.Usecases {
		buf.WriteString(fmt.Sprintf("%s.New%sUsecase,\n",
			helper.GetPkgName(usecase.Name),
			helper.GetStructName(usecase.Name),
		))
	}
	*code = strings.ReplaceAll(*code, template.PH_PROVIDER_USECASE_LIST, buf.String())
}
