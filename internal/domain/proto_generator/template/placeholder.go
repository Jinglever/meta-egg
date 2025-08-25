package template

const (
	// basic
	PH_GO_MODULE           = "%%GO-MODULE%%"
	PH_PROJECT_NAME        = "%%PROJECT-NAME%%"
	PH_PROJECT_NAME_DIR    = "%%PROJECT-NAME-DIR%%"    // 适用于目录名
	PH_PROJECT_NAME_PKG    = "%%PROJECT-NAME-PKG%%"    // 适用于包名
	PH_PROJECT_NAME_STRUCT = "%%PROJECT-NAME-STRUCT%%" // 适用于结构体名
	PH_TABLE_COMMENT       = "%%TABLE-COMMENT%%"
	PH_TABLE_NAME_URI      = "%%TABLE-NAME-URI%%"
	PH_TABLE_NAME_STRUCT   = "%%TABLE-NAME-STRUCT%%"

	// dynamic
	PH_HANDLER_MESSAGE         = "%%HANDLER-MESSAGE%%"  // 接口的入参和出参
	PH_HANDLER_FUNCTION        = "%%HANDLER-FUNCTION%%" // 接口函数定义
	PH_COL_LIST_IN_VO          = "%%COL-LIST-IN-VO%%"
	PH_COL_LIST_FOR_CREATE     = "%%COL-LIST-FOR-CREATE%%"
	PH_COL_LIST_FOR_CREATE_ADD = "%%COL-LIST-FOR-CREATE-ADD%%"
	PH_COL_LIST_FOR_LIST       = "%%COL-LIST-FOR-LIST%%"
	PH_COL_LIST_FOR_FILTER     = "%%COL-LIST-FOR-FILTER%%"
	PH_COL_LIST_FOR_ORDER      = "%%COL-LIST-FOR-ORDER%%"
	PH_COL_LIST_FOR_UPDATE     = "%%COL-LIST-FOR-UPDATE%%"
	PH_IMPORT_FOR_MESSAGE      = "%%IMPORT-FOR-MESSAGE%%"
	PH_IMPORT_TABLE_PROTO      = "%%IMPORT-TABLE-PROTO%%" // e.g. import "account.proto";

	// RL table related placeholders
	PH_RL_FIELDS_IN_DETAIL    = "%%RL-FIELDS-IN-DETAIL%%"    // RL表字段在Detail消息中
	PH_RL_FIELDS_IN_LIST      = "%%RL-FIELDS-IN-LIST%%"      // RL表字段在ListInfo消息中
	PH_RL_MESSAGES            = "%%RL-MESSAGES%%"            // RL表相关的消息定义
	PH_RL_HANDLER_FUNCTIONS   = "%%RL-HANDLER-FUNCTIONS%%"   // RL表相关的gRPC方法定义
	PH_RL_TABLE_NAME_STRUCT   = "%%RL-TABLE-NAME-STRUCT%%"   // RL表结构体名
	PH_RL_TABLE_COMMENT       = "%%RL-TABLE-COMMENT%%"       // RL表注释
	PH_MAIN_TABLE_NAME_STRUCT = "%%MAIN-TABLE-NAME-STRUCT%%" // 主表结构体名
	PH_MAIN_TABLE_COMMENT     = "%%MAIN-TABLE-COMMENT%%"     // 主表注释
	PH_MAIN_TABLE_NAME_LOWER  = "%%MAIN-TABLE-NAME-LOWER%%"  // 主表名小写
	PH_RL_TABLE_NAME_LOWER    = "%%RL-TABLE-NAME-LOWER%%"    // RL表名小写
	PH_RL_FIELDS_IN_CREATE    = "%%RL-FIELDS-IN-CREATE%%"    // RL表字段在Create请求中

	// BR table related placeholders
	PH_BR_HANDLER_FUNCTIONS      = "%%BR-HANDLER-FUNCTIONS%%"      // BR表相关的gRPC方法定义
	PH_OTHER_TABLE_NAME_STRUCT   = "%%OTHER-TABLE-NAME-STRUCT%%"   // BR关系中对方表的结构体名
	PH_OTHER_TABLE_COMMENT       = "%%OTHER-TABLE-COMMENT%%"       // BR关系中对方表的注释
	PH_TABLE_NAME_LOWER          = "%%TABLE-NAME-LOWER%%"          // 当前表名小写
	PH_OTHER_COL_LIST_FOR_FILTER = "%%OTHER-COL-LIST-FOR-FILTER%%" // BR关系中对方表的筛选字段
	PH_OTHER_COL_LIST_FOR_ORDER  = "%%OTHER-COL-LIST-FOR-ORDER%%"  // BR关系中对方表的排序字段
	PH_BR_MESSAGES               = "%%BR-MESSAGES%%"               // BR表相关的消息定义
)
