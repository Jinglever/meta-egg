package pkggen

import (
	"os"
	"path/filepath"
	"strings"

	"meta-egg/internal/domain/helper"
	"meta-egg/internal/domain/pkg_generator/template"
	"meta-egg/internal/model"

	log "github.com/sirupsen/logrus"

	jgstr "github.com/Jinglever/go-string"
)

// relativeDir2NeedConfirm: key是相对于工程根目录的路径，值是布尔值，代表是否需要确认；注意，不会递归查看子目录
func Generate(codeDir string, project *model.Project) (relativeDir2NeedConfirm map[string]bool, err error) {
	relativeDir2NeedConfirm = map[string]bool{
		filepath.Join("pkg", "log"):     true,
		filepath.Join("pkg", "gormx"):   true,
		filepath.Join("pkg", "version"): true,
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

	// pkg/log/log.go
	path = filepath.Join(codeDir, "pkg", "log", "log.go")
	err = generateGoFile(path, template.TplPkgLogLog, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate pkg/log/log.go failed: %v", err)
		return
	}

	// pkg/log/README.md
	path = filepath.Join(codeDir, "pkg", "log", "README.md")
	err = generateNonGoFile(path, template.TplPkgLogReadme, project, nil)
	if err != nil {
		log.Errorf("generate pkg/log/README.md failed: %v", err)
		return
	}

	// pkg/gormx/db.go
	path = filepath.Join(codeDir, "pkg", "gormx", "db.go")
	err = generateGoFile(path, template.TplPkgGormxDb, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate pkg/gormx/db.go failed: %v", err)
		return
	}

	// pkg/gormx/connect.go
	path = filepath.Join(codeDir, "pkg", "gormx", "connect.go")
	err = generateGoFile(path, template.TplPkgGormxConnect, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate pkg/gormx/connect.go failed: %v", err)
		return
	}

	// pkg/gormx/option.go
	path = filepath.Join(codeDir, "pkg", "gormx", "option.go")
	err = generateGoFile(path, template.TplPkgGormxOption, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate pkg/gormx/option.go failed: %v", err)
		return
	}

	// pkg/gormx/transaction.go
	path = filepath.Join(codeDir, "pkg", "gormx", "transaction.go")
	err = generateGoFile(path, template.TplPkgGormxTransaction, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate pkg/gormx/transaction.go failed: %v", err)
		return
	}

	// pkg/gormx/hook.go
	path = filepath.Join(codeDir, "pkg", "gormx", "hook.go")
	err = generateGoFile(path, template.TplPkgGormxHook, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate pkg/gormx/hook.go failed: %v", err)
		return
	}

	// pkg/version/version.go
	path = filepath.Join(codeDir, "pkg", "version", "version.go")
	err = generateGoFile(path, template.TplPkgVersionVersion, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate pkg/version/version.go failed: %v", err)
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
	code = strings.ReplaceAll(code, template.PH_GO_MODULE, project.GoModule)

	// go format
	formatted, err := jgstr.FormatGo([]byte(code))
	if err != nil {
		log.Errorf("format source failed: %v\n%s", err, code)
		return err
	}
	_, _ = f.Write(formatted)
	return nil
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

	_, _ = f.Write([]byte(code))
	return nil
}
