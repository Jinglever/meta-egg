package projgen

import (
	"os"
	"strings"

	"meta-egg/internal/domain/project_generator/template"
	"meta-egg/internal/model"

	log "github.com/sirupsen/logrus"
)

func generateReadme(path string, project *model.Project) error {
	f, err := os.Create(path)
	if err != nil {
		log.Errorf("create file failed: %v", err)
		return err
	}
	defer f.Close()
	// template
	code := template.TplReadme
	code = strings.ReplaceAll(code, template.PH_PROJECT_NAME, project.Name)
	code = strings.ReplaceAll(code, template.PH_PROJECT_DESC, project.Desc)

	_, _ = f.Write([]byte(code))
	return nil
}
