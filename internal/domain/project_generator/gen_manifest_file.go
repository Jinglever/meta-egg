package projgen

import (
	"os"
	"strings"

	"meta-egg/internal/domain/project_generator/template"
	"meta-egg/internal/model"

	log "github.com/sirupsen/logrus"
)

func generateManifestFile(path string, project *model.Project, ep ExtendParam) error {
	f, err := os.Create(path)
	if err != nil {
		log.Errorf("create file failed: %v", err)
		return err
	}
	defer f.Close()
	// template
	code := template.TplManifestFile

	if !ep.NeedDatabase {
		code = strings.ReplaceAll(code, template.PH_TPL_MANIFEST_DATABASE, "")
	} else {
		code = strings.ReplaceAll(code, template.PH_TPL_MANIFEST_DATABASE, template.TplManifestDatabase)

		if !ep.NeedTableDemo {
			code = strings.ReplaceAll(code, template.PH_TPL_MANIFEST_TABLE_DEMO, "")
		} else {
			code = strings.ReplaceAll(code, template.PH_TPL_MANIFEST_TABLE_DEMO, template.TplManifestTableDemo)
		}
	}

	code = strings.ReplaceAll(code, template.PH_PROJECT_NAME, project.Name)
	code = strings.ReplaceAll(code, template.PH_PROJECT_DESC, project.Desc)
	code = strings.ReplaceAll(code, template.PH_GO_MODULE, project.GoModule)
	code = strings.ReplaceAll(code, template.PH_GO_VERSION, project.GoVersion)
	code = strings.ReplaceAll(code, template.PH_SERVER_TYPE, string(project.ServerType))
	if project.NoAuth {
		code = strings.ReplaceAll(code, template.PH_NO_AUTH, "true")
	} else {
		code = strings.ReplaceAll(code, template.PH_NO_AUTH, "false")
	}
	code = strings.ReplaceAll(code, template.PH_DB_TYPE, string(ep.DatabaseType))
	code = strings.ReplaceAll(code, template.PH_DB_CHARSET, model.GetDefaultCharset(ep.DatabaseType))

	_, _ = f.Write([]byte(code))
	return nil
}
