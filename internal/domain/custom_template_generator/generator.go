package csttplgen

import (
	"meta-egg/internal/domain/custom_template_generator/template"
	"meta-egg/internal/domain/helper"
	"meta-egg/internal/model"
	"os"
	"path"
	"strings"

	jgstr "github.com/Jinglever/go-string"
	log "github.com/sirupsen/logrus"
)

// relativeDir2NeedConfirm: key是相对于工程根目录的路径，值是布尔值，代表是否需要确认；注意，不会递归查看子目录
func Generate(codeDir string, project *model.Project, customTemplateRoot string) (relativeDir2NeedConfirm map[string]bool, err error) {
	relativeDir2NeedConfirm = map[string]bool{}

	var tmpRD2NC map[string]bool

	tmpRD2NC, err = generate(codeDir, project, customTemplateRoot, "")
	if err != nil {
		log.Errorf("failed to generate from template dir %s: %v", customTemplateRoot, err)
		return
	}
	for k, v := range tmpRD2NC {
		relativeDir2NeedConfirm[k] = v
	}

	return
}

func generate(codeDir string, project *model.Project,
	customTemplateRoot, relativeCstTplRoot string) (relativeDir2NeedConfirm map[string]bool, err error) {
	var content []byte
	var tmpRD2NC map[string]bool
	relativeDir2NeedConfirm = map[string]bool{}

	// scan files in directory
	files, err := os.ReadDir(path.Join(customTemplateRoot, relativeCstTplRoot))
	if err != nil {
		log.Errorf("failed to read dir %s: %v", path.Join(customTemplateRoot, relativeCstTplRoot), err)
		return
	}
	for _, file := range files {
		if file.IsDir() {
			tmpRD2NC, err = generate(codeDir, project, customTemplateRoot, path.Join(relativeCstTplRoot, file.Name()))
			if err != nil {
				log.Errorf("failed to generate from template dir %s: %v", path.Join(relativeCstTplRoot, file.Name()), err)
				return
			}
			for k, v := range tmpRD2NC {
				relativeDir2NeedConfirm[k] = v
			}
		} else {
			targetDir := path.Join(codeDir, relativeCstTplRoot)
			srcDir := path.Join(customTemplateRoot, relativeCstTplRoot)

			// gen目录下的文件不需要确认
			if strings.Contains(relativeCstTplRoot, "internal/gen") {
				relativeDir2NeedConfirm[relativeCstTplRoot] = false
			} else {
				relativeDir2NeedConfirm[relativeCstTplRoot] = true
			}

			// make dir
			if err = os.MkdirAll(targetDir, 0755); err != nil {
				log.Errorf("failed to mkdir %s: %v", targetDir, err)
				return
			}

			// read template file
			filePath := path.Join(srcDir, file.Name())
			content, err = os.ReadFile(filePath)
			if err != nil {
				log.Errorf("failed to read file %s: %v", filePath, err)
				return
			}

			// replace placeholder
			content = replacePlaceHolder(content, project)

			if strings.HasSuffix(file.Name(), ".go") {
				// format go
				content, err = jgstr.FormatGo(content)
				if err != nil {
					log.Errorf("failed to format go file %s: %v", filePath, err)
					return
				}
			}

			// write file
			filePath = path.Join(targetDir, file.Name())
			err = os.WriteFile(filePath, content, 0644)
			if err != nil {
				log.Errorf("failed to write file %s: %v", filePath, err)
				return
			}
		}
	}
	return
}

func replacePlaceHolder(code []byte, project *model.Project) []byte {
	tmp := string(code)
	tmp = strings.ReplaceAll(tmp, template.PH_GO_MODULE, project.GoModule)
	tmp = strings.ReplaceAll(tmp, template.PH_PROJECT_NAME, project.Name)
	tmp = strings.ReplaceAll(tmp, template.PH_PROJECT_DESC, project.Desc)
	tmp = strings.ReplaceAll(tmp, template.PH_PROJECT_NAME_PKG, helper.GetPkgName(project.Name))
	tmp = strings.ReplaceAll(tmp, template.PH_PROJECT_NAME_DIR, helper.GetDirName(project.Name))
	tmp = strings.ReplaceAll(tmp, template.PH_PROJECT_NAME_STRUCT, helper.GetStructName(project.Name))
	return []byte(tmp)
}
