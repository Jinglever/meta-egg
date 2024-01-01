package template

import "meta-egg/internal/domain/helper"

// placeholder
const (
	PH_IMPORTS            = "%%IMPORTS%%"
	PH_CONST_COL_LIST     = "%%CONST-COL-LIST%%"
	PH_STRUCT_TABLE_NAME  = "%%STRUCT-TABLE-NAME%%"
	PH_STRUCT_COL_LIST    = "%%STRUCT-COL-LIST%%"
	PH_DB_TABLE_NAME      = "%%DB-TABLE-NAME%%"
	PH_TPL_CONST_META_IDS = "%%TPL-CONST-META-IDS%%"
	PH_CONST_META_ID_LIST = "%%CONST-META-ID-LIST%%"
	PH_TABLE_COMMENT      = "%%TABLE-COMMENT%%"
	PH_TPL_AFTER_FIND     = "%%TPL-AFTER-FIND%%"
	PH_CORRECT_TIMEZONE   = "%%CORRECT-TIMEZONE%%"
	PH_GO_MODULE          = "%%GO-MODULE%%"
)

var TplAfterFind string = `// AfterFind Fix Timezone
func (t *%%STRUCT-TABLE-NAME%%) AfterFind(tx *gorm.DB) (err error) {%%CORRECT-TIMEZONE%%
	return nil
}`

var TplConstMetaIDs string = `const (%%CONST-META-ID-LIST%%
)`

var TplTable string = helper.PH_META_EGG_HEADER + `
package model

import (
	%%IMPORTS%%
	"%%GO-MODULE%%/pkg/gormx"
	"gorm.io/gorm"
)

const (
	%%CONST-COL-LIST%%
)

// %%TABLE-COMMENT%%
type %%STRUCT-TABLE-NAME%% struct {
	%%STRUCT-COL-LIST%%
}

func (t *%%STRUCT-TABLE-NAME%%) TableName() string {
	return "%%DB-TABLE-NAME%%"
}

%%TPL-AFTER-FIND%%

%%TPL-CONST-META-IDS%%
`
