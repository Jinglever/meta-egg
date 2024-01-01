package template

const (
	// basic
	PH_GO_MODULE           = "%%GO-MODULE%%"
	PH_PROJECT_NAME        = "%%PROJECT-NAME%%"
	PH_PROJECT_NAME_DIR    = "%%PROJECT-NAME-DIR%%"    // 适用于目录名
	PH_PROJECT_NAME_PKG    = "%%PROJECT-NAME-PKG%%"    // 适用于包名
	PH_PROJECT_NAME_STRUCT = "%%PROJECT-NAME-STRUCT%%" // 适用于结构体名

	PH_TPL_RESOURCE_STRUCT_DB                  = "%%TPL-RESOURCE-STRUCT-DB%%"
	PH_TPL_RESOURCE_CONFIG_STRUCT_DB           = "%%TPL-RESOURCE-CONFIG-STRUCT-DB%%"
	PH_TPL_RESOURCE_DB                         = "%%TPL-RESOURCE-DB%%"
	PH_TPL_RESOURCE_STRUCT_ACCESS_TOKEN        = "%%TPL-RESOURCE-STRUCT-ACCESS-TOKEN%%"
	PH_TPL_RESOURCE_CONFIG_STRUCT_ACCESS_TOKEN = "%%TPL-RESOURCE-CONFIG-STRUCT-ACCESS-TOKEN%%"
	PH_TPL_RESOURCE_ACCESS_TOKEN               = "%%TPL-RESOURCE-ACCESS-TOKEN%%"
)