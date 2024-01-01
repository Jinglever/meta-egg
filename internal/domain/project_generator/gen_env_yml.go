package projgen

import (
	"os"
	"strings"

	"meta-egg/internal/domain/project_generator/template"

	log "github.com/sirupsen/logrus"
)

func generateEnvYml(path, projRoot, manifestRoot, manifestFile string) error {
	f, err := os.Create(path)
	if err != nil {
		log.Errorf("create file failed: %v", err)
		return err
	}
	defer f.Close()
	// template
	code := template.TplEnvYml
	code = strings.ReplaceAll(code, template.PH_PROJECT_ROOT, projRoot)
	code = strings.ReplaceAll(code, template.PH_MANIFEST_ROOT, manifestRoot)
	code = strings.ReplaceAll(code, template.PH_MANIFEST_FILE, manifestFile)

	_, _ = f.Write([]byte(code))
	return nil
}
