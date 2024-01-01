package projgen

import (
	"os"
	"path/filepath"
	"strings"

	"meta-egg/internal/domain/helper"
	"meta-egg/internal/domain/project_generator/template"
	"meta-egg/internal/model"

	jgfile "github.com/Jinglever/go-file"
	log "github.com/sirupsen/logrus"
)

type ExtendParam struct {
	NeedDatabase  bool
	NeedTableDemo bool
	DatabaseType  model.DatabaseType
}

func Generate(codeDir string, project *model.Project, ep ExtendParam) error {
	var err error

	projDir := codeDir
	manifestDir := filepath.Join(projDir, "_manifest")

	// 创建目录
	dirs := []string{
		manifestDir,
		filepath.Join(projDir, "build", "package"),
	}
	for _, dir := range dirs {
		if err = os.MkdirAll(dir, 0755); err != nil {
			log.Errorf("failed to mkdir %s: %v", dir, err)
			return err
		}
	}

	var path string

	// 创建.gitignore
	path = filepath.Join(projDir, ".gitignore")
	if !jgfile.IsFile(path) {
		err = generateGitIgnore(path)
		if err != nil {
			log.Errorf("generate .gitignore failed: %v", err)
			return err
		}
	}

	// 创建 readme
	path = filepath.Join(projDir, "README.md")
	if !jgfile.IsFile(path) {
		err = generateReadme(path, project)
		if err != nil {
			log.Errorf("generate readme failed: %v", err)
			return err
		}
	}

	// 创建 _manifest dtd
	path = filepath.Join(manifestDir, "meta_egg.dtd")
	if !jgfile.IsFile(path) {
		err = generateManifestDTD(path)
		if err != nil {
			log.Errorf("generate manifest dtd file failed: %v", err)
			return err
		}
	}

	// 创建 _manifest file
	manifestFile := filepath.Join(manifestDir, project.Name+".xml")
	if !jgfile.IsFile(manifestFile) {
		err = generateManifestFile(manifestFile, project, ep)
		if err != nil {
			log.Errorf("generate manifest file failed: %v", err)
			return err
		}
	}

	// 创建 env.xml
	path = filepath.Join(manifestDir, "env.yml")
	if !jgfile.IsFile(path) {
		err = generateEnvYml(path, projDir, manifestDir, manifestFile)
		if err != nil {
			log.Errorf("generate .gitignore failed: %v", err)
			return err
		}
	}

	// 创建go.mod
	path = filepath.Join(projDir, "go.mod")
	if !jgfile.IsFile(path) {
		err = generateGoMod(path, project)
		if err != nil {
			log.Errorf("generate go.mod failed: %v", err)
			return err
		}
	}

	// 创建makefile
	path = filepath.Join(projDir, "Makefile")
	if !jgfile.IsFile(path) {
		err = generateMakefile(path, project)
		if err != nil {
			log.Errorf("generate Makefile failed: %v", err)
			return err
		}
	}

	// 创建dockerfile
	path = filepath.Join(projDir, "build", "package", "Dockerfile")
	if !jgfile.IsFile(path) {
		err = generateNonGoFile(path, template.TplPackageDockerfile, project, nil)
		if err != nil {
			log.Errorf("generate build/package/Dockerfile failed: %v", err)
			return err
		}
	}

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
	replaceTplForNonGo(&code, project)

	_, _ = f.Write([]byte(code))
	return nil
}

func replaceTplForNonGo(code *string, project *model.Project) {
	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, project.GoModule)
	*code = strings.ReplaceAll(*code, template.PH_GO_VERSION, project.GoVersion)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME, project.Name)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_DESC, project.Desc)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_DIR, helper.GetDirName(project.Name))
}
