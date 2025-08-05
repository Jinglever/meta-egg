package template

import "meta-egg/internal/domain/helper"

var TplInternalRepoTableBR = helper.PH_META_EGG_HEADER + `
package repo

import (
	"context"
	"%%GO-MODULE%%/gen/model"
	gen "%%GO-MODULE%%/gen/repo"
	"%%GO-MODULE%%/internal/common/contexts"
	"%%GO-MODULE%%/internal/common/resource"
	"%%GO-MODULE%%/internal/repo/option"
	"%%GO-MODULE%%/pkg/gormx"
	jgstr "github.com/Jinglever/go-string"
)

//go:generate mockgen -package mock -destination ./mock/%%TABLE-NAME%%.go . %%TABLE-NAME-STRUCT%%Repo
type %%TABLE-NAME-STRUCT%%Repo interface {
	gen.%%TABLE-NAME-STRUCT%%Repo

%%BR-RELATION-METHODS-INTERFACE%%
}

type %%TABLE-NAME-STRUCT%%RepoImpl struct {
	gen.%%TABLE-NAME-STRUCT%%RepoImpl
}

func New%%TABLE-NAME-STRUCT%%Repo(rsrc *resource.Resource) %%TABLE-NAME-STRUCT%%Repo {
	return &%%TABLE-NAME-STRUCT%%RepoImpl{
		%%TABLE-NAME-STRUCT%%RepoImpl: gen.%%TABLE-NAME-STRUCT%%RepoImpl{
			Resource: rsrc,
		},
	}
}

%%BR-RELATION-METHODS-IMPLEMENTATION%%
`
