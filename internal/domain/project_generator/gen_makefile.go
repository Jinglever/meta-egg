package projgen

import (
	"os"
	"strings"

	"meta-egg/internal/domain/helper"
	"meta-egg/internal/domain/project_generator/template"
	"meta-egg/internal/model"

	log "github.com/sirupsen/logrus"
)

func generateMakefile(path string, project *model.Project) error {
	f, err := os.Create(path)
	if err != nil {
		log.Errorf("create file failed: %v", err)
		return err
	}
	defer f.Close()

	code := template.TplMakefile
	code = strings.ReplaceAll(code, template.PH_GO_MODULE, project.GoModule)
	code = strings.ReplaceAll(code, template.PH_PROJECT_NAME, project.Name)
	code = strings.ReplaceAll(code, template.PH_PROJECT_NAME_DIR, helper.GetDirName(project.Name))
	_, _ = f.Write([]byte(code))
	return nil
}
