package protogen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"meta-egg/internal/domain/helper"
	"meta-egg/internal/domain/proto_generator/template"
	"meta-egg/internal/model"

	log "github.com/sirupsen/logrus"
)

// relativeDir2NeedConfirm: key是相对于工程根目录的路径，值是布尔值，代表是否需要确认；注意，不会递归查看子目录
func Generate(codeDir string, project *model.Project) (relativeDir2NeedConfirm map[string]bool, err error) {
	relativeDir2NeedConfirm = map[string]bool{
		filepath.Join("api"):   true,
		filepath.Join("proto"): true,
		filepath.Join("third_party", "proto", "github.com", "envoyproxy", "protoc-gen-validate"):             false,
		filepath.Join("third_party", "proto", "github.com", "envoyproxy", "protoc-gen-validate", "validate"): false,
	}
	// 创建目录
	for dir := range relativeDir2NeedConfirm {
		path := filepath.Join(codeDir, dir)
		if err = os.MkdirAll(path, 0755); err != nil {
			log.Errorf("failed to mkdir %s: %v", dir, err)
			return
		}
	}

	var (
		path    string
		hasGRPC bool
	)
	if project.ServerType == model.ServerType_GRPC ||
		project.ServerType == model.ServerType_ALL {
		hasGRPC = true
	}

	// proto/error.proto
	path = filepath.Join(codeDir, "proto", helper.GetDirName(project.Name)+"_error.proto")
	err = generateNonGoFile(path, template.TplProtoError, project, nil, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate proto/error.proto failed: %v", err)
		return
	}

	if hasGRPC {
		// proto/table.proto
		var (
			tableWithSameNameOfProject *model.Table
			tableImportInProject       strings.Builder
		)
		if project.Database != nil {
			for _, table := range project.Database.Tables {
				if !table.HasHandler {
					continue
				}
				if helper.GetDirName(table.Name) == helper.GetDirName(project.Name) {
					// 项目名和表名相同，跳过, 留到后面处理
					tableWithSameNameOfProject = table
					continue
				}
				if table.Type == model.TableType_DATA || table.Type == model.TableType_META {
					path = filepath.Join(codeDir, "proto", helper.GetDirName(table.Name)+".proto")
					err = generateNonGoFile(path, template.TplProtoTable, project, table, helper.AddHeaderCanEdit)
					if err != nil {
						log.Errorf("generate proto/%v.proto failed: %v", helper.GetDirName(table.Name), err)
						return
					}
					tableImportInProject.WriteString(fmt.Sprintf("import \"%s.proto\";\n", helper.GetDirName(table.Name)))
				} else if table.Type == model.TableType_BR {
					// 为BR表生成独立的proto文件
					path = filepath.Join(codeDir, "proto", helper.GetDirName(table.Name)+".proto")
					err = generateNonGoFileForBRTable(path, template.TplProtoBRTable, project, table, helper.AddHeaderCanEdit)
					if err != nil {
						log.Errorf("generate proto/%v.proto failed: %v", helper.GetDirName(table.Name), err)
						return
					}
					tableImportInProject.WriteString(fmt.Sprintf("import \"%s.proto\";\n", helper.GetDirName(table.Name)))
				}
			}
		}

		path = filepath.Join(codeDir, "proto", helper.GetDirName(project.Name)+"_common.proto")
		err = generateNonGoFile(path, template.TplProtoCommon, project, nil, helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate proto/%v.proto failed: %v", helper.GetDirName(project.Name)+"_common", err)
			return
		}

		// proto/proj.proto
		path = filepath.Join(codeDir, "proto", helper.GetDirName(project.Name)+".proto")
		err = generateNonGoFileForProjProto(path, template.TplProtoProject, project,
			tableWithSameNameOfProject, tableImportInProject.String(), helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate proto/proj.proto failed: %v", err)
			return
		}
	}

	// third_party/proto/github.com/envoyproxy/protoc-gen-validate/validate/validate.proto
	path = filepath.Join(codeDir, "third_party", "proto", "github.com",
		"envoyproxy", "protoc-gen-validate", "validate", "validate.proto")
	err = generateNonGoFile(path, template.TplThirdPartyProtoValidate, project, nil, nil)
	if err != nil {
		log.Errorf("generate third_party/proto/.../validate.proto failed: %v", err)
		return
	}

	// third_party/proto/github.com/envoyproxy/protoc-gen-validate/README.md
	path = filepath.Join(codeDir, "third_party", "proto", "github.com",
		"envoyproxy", "protoc-gen-validate", "README.md")
	err = generateNonGoFile(path, template.TplThirdPartyProtoValidateReadme, project, nil, nil)
	if err != nil {
		log.Errorf("generate third_party/proto/.../proto-gen-validate/README.md failed: %v", err)
		return
	}

	return
}

func generateNonGoFile(path string, tpl string, project *model.Project, table *model.Table, addHeader func(s string) string) error {
	f, err := os.Create(path)
	if err != nil {
		log.Errorf("create file %s failed: %v", path, err)
		return err
	}
	defer f.Close()

	// template
	code := tpl
	if addHeader != nil {
		code = addHeader(code)
	}
	replaceTpl(&code, project, table)

	f.Write([]byte(code))
	return nil
}

func generateNonGoFileForProjProto(path string, tpl string, project *model.Project,
	table *model.Table, tableImports string, addHeader func(s string) string) error {
	f, err := os.Create(path)
	if err != nil {
		log.Errorf("create file %s failed: %v", path, err)
		return err
	}
	defer f.Close()

	// template
	code := tpl
	if addHeader != nil {
		code = addHeader(code)
	}
	replaceTpl(&code, project, table)
	code = strings.ReplaceAll(code, template.PH_IMPORT_TABLE_PROTO, tableImports)

	f.Write([]byte(code))
	return nil
}

func generateNonGoFileForBRTable(path string, tpl string, project *model.Project,
	table *model.Table, addHeader func(s string) string) error {
	f, err := os.Create(path)
	if err != nil {
		log.Errorf("create file %s failed: %v", path, err)
		return err
	}
	defer f.Close()

	// template
	code := tpl
	if addHeader != nil {
		code = addHeader(code)
	}
	replaceTplForBRTable(&code, project, table)

	f.Write([]byte(code))
	return nil
}

func replaceTpl(code *string, project *model.Project, table *model.Table) {
	genHandlerMessage(code, project, table)
	genHandlerFunction(code, project)

	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, project.GoModule)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME, project.Name)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_DIR, helper.GetDirName(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_PKG, helper.GetPkgName(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_STRUCT, helper.GetStructName(project.Name))
}

func replaceTplForTable(code *string, table *model.Table) {
	genColListInVO(code, table)
	genColListForCreate(code, table)
	genColListForList(code, table)

	cnt := genColListForFilter(code, table, 2)
	_ = getColListForOrder(code, table, cnt)

	_ = genColListForUpdate(code, table, 2)

	*code = strings.ReplaceAll(*code, template.PH_TABLE_COMMENT, table.Comment)
	*code = strings.ReplaceAll(*code, template.PH_TABLE_NAME_URI, helper.GetURIName(table.Name))
	*code = strings.ReplaceAll(*code, template.PH_TABLE_NAME_STRUCT, helper.GetStructName(table.Name))
}

func replaceTplForBRTable(code *string, project *model.Project, table *model.Table) {
	genBRTableHandlerMessage(code, project, table)
	// BR表不需要service函数，因为它们会被添加到主项目的service中
	*code = strings.ReplaceAll(*code, template.PH_HANDLER_FUNCTION, "")

	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, project.GoModule)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME, project.Name)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_DIR, helper.GetDirName(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_PKG, helper.GetPkgName(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_STRUCT, helper.GetStructName(project.Name))
}

// genBRTableHandlerMessage 为BR表生成proto message
func genBRTableHandlerMessage(code *string, project *model.Project, table *model.Table) {
	// 构建表名到表的映射
	tableNameToTable := make(map[string]*model.Table)
	for _, t := range project.Database.Tables {
		tableNameToTable[t.Name] = t
	}

	// 识别BR表的两个关联表
	brRelatedTables := helper.IdentifyBRRelatedTables(table, tableNameToTable)
	if brRelatedTables == nil {
		*code = strings.ReplaceAll(*code, template.PH_HANDLER_MESSAGE, "")
		return
	}

	var buf strings.Builder

	// 生成查询请求和响应message
	generateBRTableQueryMessages(&buf, table, brRelatedTables)

	// 生成批量操作请求message
	generateBRTableBatchMessages(&buf, table, brRelatedTables)

	*code = strings.ReplaceAll(*code, template.PH_HANDLER_MESSAGE, buf.String())
}

// generateBRTableQueryMessages 生成BR表查询相关的proto message
func generateBRTableQueryMessages(buf *strings.Builder, brTable *model.Table, brRelatedTables *helper.BRRelatedTables) {
	table1StructName := helper.GetStructName(brRelatedTables.Table1.Name)
	table2StructName := helper.GetStructName(brRelatedTables.Table2.Name)

	// 生成Get{Table2}ListBy{Table1}ID请求message
	buf.WriteString(fmt.Sprintf(`
message Get%sListBy%sIDRequest {
  uint64 %s_id = 1; // %sID
  uint32 page = 2; // 页码, 从1开始
  uint32 page_size = 3; // 每页数量, 要求大于0
%s%s}
`, table2StructName, table1StructName, // message名
		helper.GetVarName(brRelatedTables.Table1.Name), brRelatedTables.Table1.Comment, // 字段
		generateBRProtoFilterFields(brRelatedTables.Table2, 4), // 筛选字段
		generateBRProtoOrderFields(brRelatedTables.Table2, 4))) // 排序字段

	// 生成Get{Table1}ListBy{Table2}ID请求message
	buf.WriteString(fmt.Sprintf(`
message Get%sListBy%sIDRequest {
  uint64 %s_id = 1; // %sID
  uint32 page = 2; // 页码, 从1开始
  uint32 page_size = 3; // 每页数量, 要求大于0
%s%s}
`, table1StructName, table2StructName, // message名
		helper.GetVarName(brRelatedTables.Table2.Name), brRelatedTables.Table2.Comment, // 字段
		generateBRProtoFilterFields(brRelatedTables.Table1, 4), // 筛选字段
		generateBRProtoOrderFields(brRelatedTables.Table1, 4))) // 排序字段

	// 响应message复用主表的定义，不需要在BR表proto文件中重复定义
}

// generateBRProtoFilterFields 生成BR表proto message中的筛选字段
func generateBRProtoFilterFields(table *model.Table, startFieldNum int) string {
	var buf strings.Builder
	fieldNum := startFieldNum

	for _, col := range table.Columns {
		if !col.IsFilter || col.IsHidden {
			continue
		}

		fieldName := helper.GetVarName(col.Name)
		buf.WriteString(fmt.Sprintf("  optional string %s = %d; // %s\n", fieldName, fieldNum, col.Comment))
		fieldNum++
	}

	return buf.String()
}

// generateBRProtoOrderFields 生成BR表proto message中的排序字段
func generateBRProtoOrderFields(table *model.Table, startFieldNum int) string {
	var buf strings.Builder
	hasOrderCol := false

	for _, col := range table.Columns {
		if col.IsOrder && !col.IsHidden {
			hasOrderCol = true
			break
		}
	}

	if !hasOrderCol {
		return ""
	}

	fieldNum := startFieldNum
	// 找到筛选字段的数量，调整起始字段号
	for _, col := range table.Columns {
		if col.IsFilter && !col.IsHidden {
			fieldNum++
		}
	}

	buf.WriteString(fmt.Sprintf("  optional string order_by = %d; // 排序字段\n", fieldNum))
	buf.WriteString(fmt.Sprintf("  optional string order_type = %d; // 排序类型: asc, desc\n", fieldNum+1))

	return buf.String()
}

// generateBRTableBatchMessages 生成BR表批量操作相关的proto message
func generateBRTableBatchMessages(buf *strings.Builder, brTable *model.Table, brRelatedTables *helper.BRRelatedTables) {
	table1StructName := helper.GetStructName(brRelatedTables.Table1.Name)
	table2StructName := helper.GetStructName(brRelatedTables.Table2.Name)
	table1PluralStructName := helper.GetStructName(helper.GetPluralName(brRelatedTables.Table1.Name))
	table2PluralStructName := helper.GetStructName(helper.GetPluralName(brRelatedTables.Table2.Name))

	// 生成批量绑定请求message
	buf.WriteString(fmt.Sprintf(`
message Bind%sTo%sRequest {
  uint64 %s_id = 1; // %sID
  repeated uint64 %s_ids = 2; // %sID列表
}
`, table2PluralStructName, table1StructName, // message名
		helper.GetVarName(brRelatedTables.Table1.Name), brRelatedTables.Table1.Comment, // 字段1
		helper.GetVarName(brRelatedTables.Table2.Name), brRelatedTables.Table2.Comment)) // 字段2

	buf.WriteString(fmt.Sprintf(`
message Bind%sTo%sRequest {
  uint64 %s_id = 1; // %sID
  repeated uint64 %s_ids = 2; // %sID列表
}
`, table1PluralStructName, table2StructName, // message名
		helper.GetVarName(brRelatedTables.Table2.Name), brRelatedTables.Table2.Comment, // 字段1
		helper.GetVarName(brRelatedTables.Table1.Name), brRelatedTables.Table1.Comment)) // 字段2

	// 生成批量解绑请求message
	buf.WriteString(fmt.Sprintf(`
message Unbind%sFrom%sRequest {
  uint64 %s_id = 1; // %sID
  repeated uint64 %s_ids = 2; // %sID列表
}
`, table2PluralStructName, table1StructName, // message名
		helper.GetVarName(brRelatedTables.Table1.Name), brRelatedTables.Table1.Comment, // 字段1
		helper.GetVarName(brRelatedTables.Table2.Name), brRelatedTables.Table2.Comment)) // 字段2

	buf.WriteString(fmt.Sprintf(`
message Unbind%sFrom%sRequest {
  uint64 %s_id = 1; // %sID
  repeated uint64 %s_ids = 2; // %sID列表
}
`, table1PluralStructName, table2StructName, // message名
		helper.GetVarName(brRelatedTables.Table2.Name), brRelatedTables.Table2.Comment, // 字段1
		helper.GetVarName(brRelatedTables.Table1.Name), brRelatedTables.Table1.Comment)) // 字段2
}

// generateBRTableServiceFunctions 生成BR表的proto service函数
func generateBRTableServiceFunctions(buf *strings.Builder, brTable *model.Table, brRelatedTables *helper.BRRelatedTables) {
	table1StructName := helper.GetStructName(brRelatedTables.Table1.Name)
	table2StructName := helper.GetStructName(brRelatedTables.Table2.Name)
	table1PluralStructName := helper.GetStructName(helper.GetPluralName(brRelatedTables.Table1.Name))
	table2PluralStructName := helper.GetStructName(helper.GetPluralName(brRelatedTables.Table2.Name))

	// 生成查询service函数
	buf.WriteString(fmt.Sprintf(`    // 获取与指定%s关联的%s列表
    rpc Get%sListBy%sID (Get%sListBy%sIDRequest) returns (Get%sListResponse) {}
`, brRelatedTables.Table1.Comment, brRelatedTables.Table2.Comment, // 注释
		table2StructName, table1StructName, table2StructName, table1StructName, table2StructName)) // 函数名

	buf.WriteString(fmt.Sprintf(`    // 获取与指定%s关联的%s列表
    rpc Get%sListBy%sID (Get%sListBy%sIDRequest) returns (Get%sListResponse) {}
`, brRelatedTables.Table2.Comment, brRelatedTables.Table1.Comment, // 注释
		table1StructName, table2StructName, table1StructName, table2StructName, table1StructName)) // 函数名

	// 生成批量绑定service函数
	buf.WriteString(fmt.Sprintf(`    // 给指定%s批量关联%s
    rpc Bind%sTo%s (Bind%sTo%sRequest) returns (google.protobuf.Empty) {}
`, brRelatedTables.Table1.Comment, brRelatedTables.Table2.Comment, // 注释
		table2PluralStructName, table1StructName, table2PluralStructName, table1StructName)) // 函数名

	buf.WriteString(fmt.Sprintf(`    // 给指定%s批量关联%s
    rpc Bind%sTo%s (Bind%sTo%sRequest) returns (google.protobuf.Empty) {}
`, brRelatedTables.Table2.Comment, brRelatedTables.Table1.Comment, // 注释
		table1PluralStructName, table2StructName, table1PluralStructName, table2StructName)) // 函数名

	// 生成批量解绑service函数
	buf.WriteString(fmt.Sprintf(`    // 从指定%s解除与多个%s的关联
    rpc Unbind%sFrom%s (Unbind%sFrom%sRequest) returns (google.protobuf.Empty) {}
`, brRelatedTables.Table1.Comment, brRelatedTables.Table2.Comment, // 注释
		table2PluralStructName, table1StructName, table2PluralStructName, table1StructName)) // 函数名

	buf.WriteString(fmt.Sprintf(`    // 从指定%s解除与多个%s的关联
    rpc Unbind%sFrom%s (Unbind%sFrom%sRequest) returns (google.protobuf.Empty) {}
`, brRelatedTables.Table2.Comment, brRelatedTables.Table1.Comment, // 注释
		table1PluralStructName, table2StructName, table1PluralStructName, table2StructName)) // 函数名
}

// genBRTableServiceFunctions 为BR表生成service函数（用于添加到主service中）
func genBRTableServiceFunctions(brTable *model.Table, project *model.Project) string {
	// 构建表名到表的映射
	tableNameToTable := make(map[string]*model.Table)
	for _, t := range project.Database.Tables {
		tableNameToTable[t.Name] = t
	}

	// 识别BR表的两个关联表
	brRelatedTables := helper.IdentifyBRRelatedTables(brTable, tableNameToTable)
	if brRelatedTables == nil {
		return ""
	}

	var buf strings.Builder

	// 生成service函数
	generateBRTableServiceFunctions(&buf, brTable, brRelatedTables)

	return buf.String()
}

func genColListInVO(code *string, table *model.Table) {
	var (
		buf     strings.Builder
		colType string
		err     error
	)
	cnt := 1
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}

		colType, err = helper.GetProto3ValueType(col)
		if err != nil {
			log.Fatalf("fail to get to type: %v", err)
		}
		if !col.IsRequired {
			colType = "optional " + colType
		}
		buf.WriteString(fmt.Sprintf("    %s %s = %d; // %s\n",
			colType,
			helper.GetDirName(col.Name),
			cnt,
			col.Comment))
		cnt++
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_IN_VO, buf.String())
}

func genColListForList(code *string, table *model.Table) {
	var (
		buf     strings.Builder
		colType string
		err     error
	)
	cnt := 1
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}
		if !col.IsList {
			continue
		}

		colType, err = helper.GetProto3ValueType(col)
		if err != nil {
			log.Fatalf("fail to get to type: %v", err)
		}
		if !col.IsRequired {
			colType = "optional " + colType
		}
		buf.WriteString(fmt.Sprintf("    %s %s = %d; // %s\n",
			colType,
			helper.GetDirName(col.Name),
			cnt,
			col.Comment))
		cnt++
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_LIST, buf.String())
}

func genColListForCreate(code *string, table *model.Table) {
	var (
		buf     strings.Builder
		colType string
		err     error
	)
	cnt := 1
	for _, col := range table.Columns {
		if !col.IsAlterable {
			continue
		}
		if col.IsHidden {
			continue
		}

		colType, err = helper.GetProto3ValueType(col)
		if err != nil {
			log.Fatalf("fail to get to type: %v", err)
		}
		if !col.IsRequired {
			colType = "optional " + colType
		}
		// comment
		buf.WriteString(fmt.Sprintf("    // %s\n",
			helper.GetCommentForHandler(col)))
		buf.WriteString(fmt.Sprintf("    %s %s = %d%s;\n",
			colType,
			helper.GetDirName(col.Name),
			cnt,
			helper.GetProto3ValidateRule(col),
		))
		cnt++
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_CREATE, buf.String())
}

func genHandlerMessage(code *string, project *model.Project, table *model.Table) {
	if project.ServerType != model.ServerType_GRPC &&
		project.ServerType != model.ServerType_ALL {
		return
	}
	if table == nil {
		// import
		*code = strings.ReplaceAll(*code, template.PH_IMPORT_FOR_MESSAGE, "")
		*code = strings.ReplaceAll(*code, template.PH_HANDLER_MESSAGE, "")
		return
	} else {
		// import
		*code = strings.ReplaceAll(*code, template.PH_IMPORT_FOR_MESSAGE, template.TplProtoImportForMessage)
	}

	// message
	var buf strings.Builder
	if table.Type == model.TableType_DATA {
		str := template.TplProtoDataTableMessage
		replaceTplForTable(&str, table)
		// 添加RL表支持
		genRLFieldsAndMessages(&str, table, project)
		// 添加BR表支持
		genBRMessages(&str, table, project)
		buf.WriteString(str)
	} else if table.Type == model.TableType_META {
		str := template.TplProtoMetaTableMessage
		replaceTplForTable(&str, table)
		// META表不需要BR表支持
		genBRMessages(&str, table, project) // 会清空BR占位符
		buf.WriteString(str)
	}
	*code = strings.ReplaceAll(*code, template.PH_HANDLER_MESSAGE, buf.String())
}

func genHandlerFunction(code *string, project *model.Project) {
	if project.ServerType != model.ServerType_GRPC &&
		project.ServerType != model.ServerType_ALL {
		return
	}
	var buf strings.Builder

	if project.Database != nil {
		for _, table := range project.Database.Tables {
			if !table.HasHandler {
				continue
			}
			if table.Type == model.TableType_DATA {
				str := template.TplProtoDataTableHandlerFuncs
				replaceTplForTable(&str, table)
				// 添加RL表gRPC方法
				genRLHandlerFunctions(&str, table, project)
				// 添加BR表gRPC方法
				genBRHandlerFunctions(&str, table, project)
				buf.WriteString(str)
			} else if table.Type == model.TableType_META {
				str := template.TplProtoMetaTableHandlerFuncs
				replaceTplForTable(&str, table)
				// META表不需要BR表支持
				genBRHandlerFunctions(&str, table, project) // 会清空BR占位符
				buf.WriteString(str)
			} else if table.Type == model.TableType_BR {
				// 为BR表生成service函数并添加到主service中
				str := genBRTableServiceFunctions(table, project)
				buf.WriteString(str)
			}
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_HANDLER_FUNCTION, buf.String())
}

func genColListForFilter(code *string, table *model.Table, cnt int) int {
	var (
		buf     strings.Builder
		colType string
		err     error
	)
	hasFilterCol := false
	for _, col := range table.Columns {
		if col.IsFilter && !col.IsHidden {
			hasFilterCol = true
			break
		}
	}
	if !hasFilterCol {
		*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_FILTER, "")
		return cnt
	}
	for _, col := range table.Columns {
		if !col.IsFilter || col.IsHidden {
			continue
		}

		colType, err = helper.GetProto3ValueType(col)
		if err != nil {
			log.Fatalf("fail to get to type: %v", err)
		}
		colType = "optional " + colType
		// comment
		buf.WriteString(fmt.Sprintf("    // 筛选项: %s (可选)\n",
			helper.GetCommentForHandler(col)))
		buf.WriteString(fmt.Sprintf("    %s %s = %d%s;\n",
			colType,
			helper.GetDirName(col.Name),
			cnt,
			helper.GetProto3ValidateRule(col),
		))
		cnt++
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_FILTER, buf.String())
	return cnt
}

func getColListForOrder(code *string, table *model.Table, cnt int) int {
	var buf strings.Builder
	hasOrderCol := false
	for _, col := range table.Columns {
		if col.IsOrder && !col.IsHidden {
			hasOrderCol = true
			break
		}
	}
	if !hasOrderCol {
		*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_ORDER, "")
		return cnt
	}
	orderCols := make([]string, 0)
	for _, col := range table.Columns {
		if !col.IsOrder || col.IsHidden {
			continue
		}
		orderCols = append(orderCols, col.Name)
	}
	buf.WriteString(fmt.Sprintf("    // 排序字段, 可选: %s\n",
		strings.Join(orderCols, ", ")))
	buf.WriteString(fmt.Sprintf("    optional string order_by = %d [(validate.rules).string = {in: [\"%s\"]}];\n",
		cnt,
		strings.Join(orderCols, "\", \""),
	))
	cnt++
	buf.WriteString("    // 排序方式, 默认 desc, 可选: asc, desc\n")
	buf.WriteString(fmt.Sprintf("    optional string order_type = %d [(validate.rules).string = {in: [\"asc\", \"desc\"]}];\n", cnt))
	cnt++
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_ORDER, buf.String())
	return cnt
}

func genColListForUpdate(code *string, table *model.Table, cnt int) int {
	var (
		buf     strings.Builder
		colType string
		err     error
	)
	for _, col := range table.Columns {
		if !col.IsAlterable {
			continue
		}
		if col.IsHidden {
			continue
		}

		colType, err = helper.GetProto3ValueType(col)
		if err != nil {
			log.Fatalf("fail to get to type: %v", err)
		}
		colType = "optional " + colType
		// comment
		buf.WriteString(fmt.Sprintf("    // 更新项: %s (可选)\n",
			helper.GetCommentForHandler(col)))
		buf.WriteString(fmt.Sprintf("    %s %s = %d%s;\n",
			colType,
			helper.GetDirName(col.Name),
			cnt,
			helper.GetProto3ValidateRule(col),
		))
		cnt++
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_UPDATE, buf.String())
	return cnt
}

func genRLFieldsAndMessages(code *string, table *model.Table, project *model.Project) {
	if project.Database == nil {
		// 清空占位符
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_DETAIL, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_LIST, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_MESSAGES, "")
		return
	}

	// 创建表名到表的映射
	tableNameToTable := make(map[string]*model.Table)
	for _, t := range project.Database.Tables {
		tableNameToTable[t.Name] = t
	}

	// 获取该主表的所有RL表
	rlTables := helper.GetMainTableRLs(table, project.Database.Tables)

	var (
		detailFieldsBuf strings.Builder
		listFieldsBuf   strings.Builder
		createFieldsBuf strings.Builder
		messagesBuf     strings.Builder
		fieldIndex      = getLastFieldIndex(table) + 1 // 从主表字段后继续编号
	)

	for _, rlTable := range rlTables {
		// 为Detail消息添加RL表字段（所有RL表都包含）
		detailFieldsBuf.WriteString(fmt.Sprintf("    repeated %sDetail %s = %d; // %s列表\n",
			helper.GetStructName(rlTable.Name),
			helper.GetDirName(rlTable.Name)+"s",
			fieldIndex,
			rlTable.Comment))
		detailFieldIndex := fieldIndex
		fieldIndex++

		// 为ListInfo消息添加RL表字段（仅包含有list=true字段的RL表）
		if helper.ShouldIncludeRLInList(rlTable) {
			listFieldsBuf.WriteString(fmt.Sprintf("    repeated %sListInfo %s = %d; // %s列表\n",
				helper.GetStructName(rlTable.Name),
				helper.GetDirName(rlTable.Name)+"s",
				detailFieldIndex, // 使用相同的字段索引
				rlTable.Comment))
		}

		// 为Create请求添加RL表字段（包含alter=true的字段）
		createFieldsBuf.WriteString(genRLFieldsForCreate(rlTable, fieldIndex))
		fieldIndex++

		// 生成RL表相关消息
		rlMsg := template.TplRLTableMessage
		replaceTplForRLTable(&rlMsg, rlTable, table)
		messagesBuf.WriteString(rlMsg)
	}

	// 替换占位符
	*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_DETAIL, detailFieldsBuf.String())
	*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_LIST, listFieldsBuf.String())
	*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_CREATE, createFieldsBuf.String())
	*code = strings.ReplaceAll(*code, template.PH_RL_MESSAGES, messagesBuf.String())
}

func genRLHandlerFunctions(code *string, table *model.Table, project *model.Project) {
	if project.Database == nil {
		// 清空占位符
		*code = strings.ReplaceAll(*code, template.PH_RL_HANDLER_FUNCTIONS, "")
		return
	}

	// 获取该主表的所有RL表
	rlTables := helper.GetMainTableRLs(table, project.Database.Tables)

	var handlerFuncsBuf strings.Builder

	for _, rlTable := range rlTables {
		// 生成RL表的gRPC方法
		rlHandlerFunc := template.TplRLTableHandlerFuncs
		replaceTplForRLTable(&rlHandlerFunc, rlTable, table)
		handlerFuncsBuf.WriteString(rlHandlerFunc)
	}

	// 替换占位符
	*code = strings.ReplaceAll(*code, template.PH_RL_HANDLER_FUNCTIONS, handlerFuncsBuf.String())
}

// getLastFieldIndex 获取表中最后一个字段的索引（用于proto字段编号）
func getLastFieldIndex(table *model.Table) int {
	maxIndex := 0
	for _, col := range table.Columns {
		if !col.IsHidden {
			maxIndex++
		}
	}
	return maxIndex
}

// replaceTplForRLTable 为RL表模板替换占位符
func replaceTplForRLTable(code *string, rlTable *model.Table, mainTable *model.Table) {
	// 替换RL表相关占位符
	*code = strings.ReplaceAll(*code, template.PH_RL_TABLE_NAME_STRUCT, helper.GetStructName(rlTable.Name))
	*code = strings.ReplaceAll(*code, template.PH_RL_TABLE_COMMENT, rlTable.Comment)
	*code = strings.ReplaceAll(*code, template.PH_MAIN_TABLE_NAME_STRUCT, helper.GetStructName(mainTable.Name))
	*code = strings.ReplaceAll(*code, template.PH_MAIN_TABLE_COMMENT, mainTable.Comment)
	*code = strings.ReplaceAll(*code, template.PH_MAIN_TABLE_NAME_LOWER, helper.GetDirName(mainTable.Name))
	*code = strings.ReplaceAll(*code, template.PH_RL_TABLE_NAME_LOWER, helper.GetDirName(rlTable.Name))

	// 生成RL表的字段列表
	genColListInVO(code, rlTable)
	genColListForCreate(code, rlTable)
	genColListForList(code, rlTable)
}

func genRLFieldsForCreate(rlTable *model.Table, startFieldIndex int) string {
	var buf strings.Builder
	fieldIndex := startFieldIndex

	// 只生成alter=true的字段，以repeated形式包含在Create请求中
	hasAlterableFields := false
	for _, col := range rlTable.Columns {
		if col.IsAlterable && !col.IsHidden {
			hasAlterableFields = true
			break
		}
	}

	if hasAlterableFields {
		// 生成RL表的创建数据结构
		buf.WriteString(fmt.Sprintf("    // %s数据列表（可选）\n", rlTable.Comment))
		buf.WriteString(fmt.Sprintf("    repeated %sCreateData %s = %d;\n",
			helper.GetStructName(rlTable.Name),
			helper.GetDirName(rlTable.Name)+"s",
			fieldIndex))
	}

	return buf.String()
}

// genBRMessages 为DATA表生成BR关系的protobuf消息（已弃用，BR表现在有独立的proto文件）
func genBRMessages(code *string, table *model.Table, project *model.Project) {
	// BR表现在有独立的proto文件，不再在主表proto中生成BR相关message
	*code = strings.ReplaceAll(*code, template.PH_BR_MESSAGES, "")
}

// genBRHandlerFunctions 为DATA表生成BR关系的gRPC方法（已弃用，BR表现在有独立的service函数）
func genBRHandlerFunctions(code *string, table *model.Table, project *model.Project) {
	// BR表现在有独立的service函数，不再在主表service中生成BR相关方法
	*code = strings.ReplaceAll(*code, template.PH_BR_HANDLER_FUNCTIONS, "")
}

// genOtherColListForFilter 生成对方表的筛选字段（protobuf格式）
func genOtherColListForFilter(code *string, table *model.Table, startCnt int) int {
	var buf strings.Builder
	cnt := startCnt

	// 检查是否有筛选字段
	hasFilterCol := false
	for _, col := range table.Columns {
		if col.IsFilter && !col.IsHidden {
			hasFilterCol = true
			break
		}
	}

	if !hasFilterCol {
		*code = strings.ReplaceAll(*code, template.PH_OTHER_COL_LIST_FOR_FILTER, "")
		return cnt
	}

	for _, col := range table.Columns {
		if !col.IsFilter || col.IsHidden {
			continue
		}
		colType, err := helper.GetProto3ValueType(col)
		if err != nil {
			log.Fatalf("fail to get to type: %v", err)
		}
		colType = "optional " + colType
		// comment
		buf.WriteString(fmt.Sprintf("    // 筛选条件: %s (可选)\n", helper.GetCommentForHandler(col)))
		buf.WriteString(fmt.Sprintf("    %s %s = %d%s;\n",
			colType,
			helper.GetDirName(col.Name),
			cnt,
			helper.GetProto3ValidateRule(col),
		))
		cnt++
	}

	*code = strings.ReplaceAll(*code, template.PH_OTHER_COL_LIST_FOR_FILTER, buf.String())
	return cnt
}

// genOtherColListForOrder 生成对方表的排序字段（protobuf格式）
func genOtherColListForOrder(code *string, table *model.Table, startCnt int) int {
	var buf strings.Builder
	cnt := startCnt

	// 检查是否有排序字段
	hasOrderCol := false
	for _, col := range table.Columns {
		if col.IsOrder && !col.IsHidden {
			hasOrderCol = true
			break
		}
	}

	if !hasOrderCol {
		*code = strings.ReplaceAll(*code, template.PH_OTHER_COL_LIST_FOR_ORDER, "")
		return cnt
	}

	orderCols := make([]string, 0)
	for _, col := range table.Columns {
		if !col.IsOrder || col.IsHidden {
			continue
		}
		orderCols = append(orderCols, col.Name)
	}
	buf.WriteString(fmt.Sprintf("    // 排序字段, 可选: %s\n",
		strings.Join(orderCols, ", ")))
	buf.WriteString(fmt.Sprintf("    optional string order_by = %d [(validate.rules).string = {in: [\"%s\"]}];\n",
		cnt,
		strings.Join(orderCols, "\", \""),
	))
	cnt++
	buf.WriteString("    // 排序方式, 默认 desc, 可选: asc, desc\n")
	buf.WriteString(fmt.Sprintf("    optional string order_type = %d [(validate.rules).string = {in: [\"asc\", \"desc\"]}];\n", cnt))
	cnt++

	*code = strings.ReplaceAll(*code, template.PH_OTHER_COL_LIST_FOR_ORDER, buf.String())
	return cnt
}
