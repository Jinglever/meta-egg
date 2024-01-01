package modeler

import (
	"encoding/xml"
	"os"

	"meta-egg/internal/domain/helper"
	"meta-egg/internal/model"

	jgstr "github.com/Jinglever/go-string"
	log "github.com/sirupsen/logrus"
)

type Modeler struct {
	Project *model.Project
}

// NewModeler create a new modeler from xml file
func ParseXMLFile(path string) (*Modeler, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Errorf("open xml file failed: %v", err)
		return nil, err
	}
	defer f.Close()

	var proj model.Project
	d := xml.NewDecoder(f)
	err = d.Decode(&proj)
	if err != nil {
		log.Errorf("decode xml file failed: %v", err)
		return nil, err
	}
	proj.Name = helper.NormalizeProjectName(proj.Name)
	if err := proj.Validate(); err != nil {
		log.Errorf("validate project failed: %v", err)
		return nil, err
	}
	proj.MakeUp()
	log.Debugf("project: %s", jgstr.JsonEncode(proj))
	return &Modeler{Project: &proj}, nil
}

// to json
func (m *Modeler) ToJSON() string {
	return jgstr.JsonEncode(m.Project)
}
