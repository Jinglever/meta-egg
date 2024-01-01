package projgen

import (
	"os"

	"meta-egg/internal/domain/project_generator/template"

	log "github.com/sirupsen/logrus"
)

func generateManifestDTD(path string) error {
	f, err := os.Create(path)
	if err != nil {
		log.Errorf("create file failed: %v", err)
		return err
	}
	defer f.Close()
	_, _ = f.Write([]byte(template.TplManifestDTD))
	return nil
}
