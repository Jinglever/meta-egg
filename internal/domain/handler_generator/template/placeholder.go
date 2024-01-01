package template

const (
	// basic
	PH_GO_MODULE           = "%%GO-MODULE%%"
	PH_GO_VERSION          = "%%GO-VERSION%%"
	PH_ENV_PREFIX          = "%%ENV-PREFIX%%"
	PH_PROJECT_NAME        = "%%PROJECT-NAME%%"
	PH_PROJECT_NAME_DIR    = "%%PROJECT-NAME-DIR%%"    // 适用于目录名
	PH_PROJECT_NAME_PKG    = "%%PROJECT-NAME-PKG%%"    // 适用于包名
	PH_PROJECT_NAME_STRUCT = "%%PROJECT-NAME-STRUCT%%" // 适用于结构体名
	PH_TABLE_COMMENT       = "%%TABLE-COMMENT%%"
	PH_TABLE_NAME_STRUCT   = "%%TABLE-NAME-STRUCT%%"
	PH_TABLE_NAME_URI      = "%%TABLE-NAME-URI%%"
	PH_TABLE_NAME          = "%%TABLE-NAME%%"
	PH_TABLE_NAME_VAR      = "%%TABLE-NAME-VAR%%" // 适用于变量名, 驼峰, 首字母小写
	PH_COMMENT_DOMAIN      = "%%COMMENT-DOMAIN%%"
	PH_COMMENT_REPO        = "%%COMMENT-REPO%%"

	// dynamic generate
	PH_USECASE_LIST_IN_STRUCT          = "%%USECASE-LIST-IN-STRUCT%%" // 在结构体中
	PH_USECASE_LIST_IN_ARG             = "%%USECASE-LIST-IN-ARG%%"    // 在函数参数中
	PH_ASSIGN_USECASE_LIST             = "%%ASSIGN-USECASE-LIST%%"    // 为结构体赋值
	PH_IMPORT_USECASE_LIST             = "%%IMPORT-USECASE-LIST%%"
	PH_ASSIGN_MODEL_TO_VO              = "%%ASSIGN-MODEL-TO-VO%%"
	PH_ASSIGN_MODEL_TO_VO_GRPC         = "%%ASSIGN-MODEL-TO-VO-GRPC%%"
	PH_COL_LIST_IN_VO                  = "%%COL-LIST-IN-VO%%"
	PH_COL_LIST_FOR_CREATE             = "%%COL-LIST-FOR-CREATE%%"
	PH_ASSIGN_CREATE_TO_MODEL          = "%%ASSIGN-CREATE-TO-MODEL%%"
	PH_COL_LIST_FOR_FILTER             = "%%COL-LIST-FOR-FILTER%%"
	PH_COL_LIST_FOR_ORDER              = "%%COL-LIST-FOR-ORDER%%"
	PH_COL_LIST_FOR_FILTER_DOC         = "%%COL-LIST-FOR-FILTER-DOC%%"
	PH_COL_LIST_FOR_ORDER_DOC          = "%%COL-LIST-FOR-ORDER-DOC%%"
	PH_PREPARE_ASSIGN_FILTER_TO_OPTION = "%%PREPARE-ASSIGN-FILTER-TO-OPTION%%"
	PH_ASSIGN_FILTER_TO_OPTION         = "%%ASSIGN-FILTER-TO-OPTION%%"
	PH_ASSIGN_ORDER_TO_OPTION          = "%%ASSIGN-ORDER-TO-OPTION%%"
	PH_COL_LIST_FOR_LIST               = "%%COL-LIST-FOR-LIST%%"
	PH_ASSIGN_MODEL_FOR_LIST           = "%%ASSIGN-MODEL-FOR-LIST%%"
	PH_COL_LIST_TO_SELECT_FOR_LIST     = "%%COL-LIST-TO-SELECT-FOR-LIST%%"
	PH_COL_LIST_FOR_UPDATE             = "%%COL-LIST-FOR-UPDATE%%"
	PH_PREPARE_ASSIGN_UPDATE_TO_SET    = "%%PREPARE-ASSIGN-UPDATE-TO-SET%%"
	PH_ASSIGN_UPDATE_TO_SET            = "%%ASSIGN-UPDATE-TO-SET%%"
	PH_PREPARE_ASSIGN_MODEL_TO_VO      = "%%PREPARE-ASSIGN-MODEL-TO-VO%%"
	PH_PREPARE_ASSIGN_MODEL_FOR_LIST   = "%%PREPARE-ASSIGN-MODEL-FOR-LIST%%"

	PH_PREPARE_ASSIGN_BO_TO_VO              = "%%PREPARE-ASSIGN-BO-TO-VO%%"
	PH_ASSIGN_BO_TO_VO                      = "%%ASSIGN-BO-TO-VO%%"
	PH_ASSIGN_BO_TO_VO_GRPC                 = "%%ASSIGN-BO-TO-VO-GRPC%%"
	PH_PREPARE_ASSIGN_CREATE_TO_BO          = "%%PREPARE-ASSIGN-CREATE-TO-BO%%"
	PH_ASSIGN_CREATE_TO_BO                  = "%%ASSIGN-CREATE-TO-BO%%"
	PH_PREPARE_ASSIGN_CREATE_TO_BO_GRPC     = "%%PREPARE-ASSIGN-CREATE-TO-BO-GRPC%%"
	PH_ASSIGN_CREATE_TO_BO_GRPC             = "%%ASSIGN-CREATE-TO-BO-GRPC%%"
	PH_PREPARE_ASSIGN_BO_FOR_LIST           = "%%PREPARE-ASSIGN-BO-FOR-LIST%%"
	PH_ASSIGN_BO_FOR_LIST                   = "%%ASSIGN-BO-FOR-LIST%%"
	PH_PREPARE_ASSIGN_FILTER_TO_OPTION_GRPC = "%%PREPARE-ASSIGN-FILTER-TO-OPTION-GRPC%%"
	PH_ASSIGN_FILTER_TO_OPTION_GRPC         = "%%ASSIGN-FILTER-TO-OPTION-GRPC%%"
	PH_PREPARE_ASSIGN_UPDATE_TO_SET_GRPC    = "%%PREPARE-ASSIGN-UPDATE-TO-SET-GRPC%%"
	PH_ASSIGN_UPDATE_TO_SET_GRPC            = "%%ASSIGN-UPDATE-TO-SET-GRPC%%"

	PH_TPL_GRPC_HANDLER_CREATE   = "%%TPL-GRPC-HANDLER-CREATE%%"
	PH_TPL_GRPC_HANDLER_GET_LIST = "%%TPL-GRPC-HANDLER-GET-LIST%%"
	PH_TPL_GRPC_HANDLER_UPDATE   = "%%TPL-GRPC-HANDLER-UPDATE%%"
	PH_TPL_GRPC_HANDLER_DELETE   = "%%TPL-GRPC-HANDLER-DELETE%%"

	PH_TPL_HTTP_HANDLER_CREATE   = "%%TPL-HTTP-HANDLER-CREATE%%"
	PH_TPL_HTTP_HANDLER_GET_LIST = "%%TPL-HTTP-HANDLER-GET-LIST%%"
	PH_TPL_HTTP_HANDLER_UPDATE   = "%%TPL-HTTP-HANDLER-UPDATE%%"
	PH_TPL_HTTP_HANDLER_DELETE   = "%%TPL-HTTP-HANDLER-DELETE%%"
)