package handlergen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"meta-egg/internal/domain/handler_generator/template"
	"meta-egg/internal/domain/helper"
	"meta-egg/internal/model"

	log "github.com/sirupsen/logrus"

	jgstr "github.com/Jinglever/go-string"
)

// relativeDir2NeedConfirm: key是相对于工程根目录的路径，值是布尔值，代表是否需要确认；注意，不会递归查看子目录
func Generate(codeDir string, project *model.Project) (relativeDir2NeedConfirm map[string]bool, err error) {
	var (
		hasGRPC bool
		hasHTTP bool
	)
	if project.ServerType == model.ServerType_GRPC {
		hasGRPC = true
	} else if project.ServerType == model.ServerType_HTTP {
		hasHTTP = true
	} else {
		hasGRPC = true
		hasHTTP = true
	}
	relativeDir2NeedConfirm = make(map[string]bool)
	if hasGRPC {
		relativeDir2NeedConfirm[filepath.Join("internal", "handler", "grpc")] = true
	}
	if hasHTTP {
		relativeDir2NeedConfirm[filepath.Join("internal", "handler", "http")] = true
	}
	// 创建目录
	for dir := range relativeDir2NeedConfirm {
		path := filepath.Join(codeDir, dir)
		if err = os.MkdirAll(path, 0755); err != nil {
			log.Errorf("failed to mkdir %s: %v", dir, err)
			return
		}
	}

	var path string

	if hasGRPC {
		// internal/handler/grpc/base.go
		path = filepath.Join(codeDir, "internal", "handler", "grpc", "base.go")
		err = generateGoFile(path, template.TplGRPCBase, project, helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate internal/handler/grpc/base.go failed: %v", err)
			return
		}

		// internal/handler/grpc/wire_gen.go
		path = filepath.Join(codeDir, "internal", "handler", "grpc", "wire_gen.go")
		err = generateGoFile(path, template.TplInternalHandlerGRPCWireGen, project, helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate internal/handler/grpc/wire_gen.go failed: %v", err)
			return
		}

		// internal/handler/grpc/wire.go
		path = filepath.Join(codeDir, "internal", "handler", "grpc", "wire.go")
		err = generateGoFile(path, template.TplInternalHandlerGRPCWire, project, helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate internal/handler/grpc/wire.go failed: %v", err)
			return
		}
	}

	if hasHTTP {
		// internal/handler/http/base.go
		path = filepath.Join(codeDir, "internal", "handler", "http", "base.go")
		err = generateGoFile(path, template.TplHTTPBase, project, helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate internal/handler/http/base.go failed: %v", err)
			return
		}

		// internal/handler/http/wire_gen.go
		path = filepath.Join(codeDir, "internal", "handler", "http", "wire_gen.go")
		err = generateGoFile(path, template.TplInternalHandlerHTTPWireGen, project, helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate internal/handler/http/wire_gen.go failed: %v", err)
			return
		}

		// internal/handler/http/wire.go
		path = filepath.Join(codeDir, "internal", "handler", "http", "wire.go")
		err = generateGoFile(path, template.TplInternalHandlerHTTPWire, project, helper.AddHeaderCanEdit)
		if err != nil {
			log.Errorf("generate internal/handler/http/wire.go failed: %v", err)
			return
		}
	}

	if project.Database != nil {
		// 按照表生成文件
		for _, table := range project.Database.Tables {
			if !table.HasHandler {
				continue
			}
			if hasGRPC {
				if table.Type == model.TableType_DATA || table.Type == model.TableType_META {
					// internal/handler/grpc/<table>.go
					path = filepath.Join(codeDir, "internal", "handler", "grpc", helper.GetDirName(table.Name)+".go")
					err = generateGoFileForTable(path, template.TplGRPCTable, table, helper.AddHeaderCanEdit)
					if err != nil {
						log.Errorf("generate internal/handler/grpc/%s.go failed: %v", helper.GetDirName(table.Name), err)
						return
					}
				}
			}

			if hasHTTP {
				if table.Type == model.TableType_DATA || table.Type == model.TableType_META {
					// internal/handler/http/<table>.go
					path = filepath.Join(codeDir, "internal", "handler", "http", helper.GetDirName(table.Name)+".go")
					err = generateGoFileForTable(path, template.TplHTTPDataTable, table, helper.AddHeaderCanEdit)
					if err != nil {
						log.Errorf("generate internal/handler/http/%s.go failed: %v", helper.GetDirName(table.Name), err)
						return
					}
				}
			}
		}
	}
	return
}

func generateGoFileForTable(path string, tpl string, table *model.Table, addHeader func(s string) string) error {
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
	replaceTplForTable(&code, table)

	// go format
	formatted, err := jgstr.FormatGo([]byte(code))
	if err != nil {
		log.Errorf("format source failed: %v\n%s", err, code)
		return err
	}
	_, _ = f.Write(formatted)
	return nil
}

func replaceTplForTable(code *string, table *model.Table) {
	if table.Type == model.TableType_DATA {
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_CREATE, template.TplGRPCHandlerCreate)
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_GET_LIST, template.TplGRPCHandlerGetList)
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_UPDATE, template.TplGRPCHandlerUpdate)
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_DELETE, template.TplGRPCHandlerDelete)

		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_CREATE, template.TplHTTPHandlerCreate)
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_GET_LIST, template.TplHTTPHandlerGetList)
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_UPDATE, template.TplHTTPHandlerUpdate)
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_DELETE, template.TplHTTPHandlerDelete)

		// 生成RL表相关内容
		genRLDetailStructsAndFields(code, table)
		genRLGRPCHandlerFunctions(code, table)
		// 生成BR表相关内容
		genBRHTTPHandlerFunctions(code, table)
		genBRGRPCHandlerFunctions(code, table)
	} else if table.Type == model.TableType_META {
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_GET_LIST, template.TplGRPCHandlerGetList)
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_UPDATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_DELETE, "")

		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_GET_LIST, template.TplHTTPHandlerGetList)
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_UPDATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_DELETE, "")

		// META表不需要RL表支持
		*code = strings.ReplaceAll(*code, template.PH_RL_DETAIL_STRUCTS, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_REQUEST_STRUCTS, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_LISTINFO_STRUCTS, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_DETAIL, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_CONVERT_IN_TO_DETAIL, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_ASSIGN_IN_DETAIL, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_CREATE_ASSIGN_TO_BO, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_LISTINFO, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_CONVERT_IN_TO_LISTINFO, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_ASSIGN_IN_LISTINFO, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_HANDLER_FUNCTIONS, "")
		// gRPC RL表相关占位符清理
		*code = strings.ReplaceAll(*code, template.PH_RL_CONVERT_IN_TO_DETAIL_GRPC, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_ASSIGN_IN_DETAIL_GRPC, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_CREATE_ASSIGN_TO_BO_GRPC, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_CONVERT_IN_TO_LISTINFO_GRPC, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_ASSIGN_IN_LISTINFO_GRPC, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_GRPC_HANDLER_FUNCTIONS, "")
		// META表不需要BR表支持
		*code = strings.ReplaceAll(*code, template.PH_BR_HTTP_HANDLER_FUNCTIONS, "")
		*code = strings.ReplaceAll(*code, template.PH_BR_GRPC_HANDLER_FUNCTIONS, "")
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_GET_LIST, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_UPDATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_DELETE, "")

		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_GET_LIST, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_UPDATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_DELETE, "")

		// 其他表类型不需要RL表支持
		*code = strings.ReplaceAll(*code, template.PH_RL_DETAIL_STRUCTS, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_REQUEST_STRUCTS, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_LISTINFO_STRUCTS, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_DETAIL, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_CONVERT_IN_TO_DETAIL, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_ASSIGN_IN_DETAIL, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_CREATE_ASSIGN_TO_BO, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_LISTINFO, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_CONVERT_IN_TO_LISTINFO, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_ASSIGN_IN_LISTINFO, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_HANDLER_FUNCTIONS, "")
		// gRPC RL表相关占位符清理
		*code = strings.ReplaceAll(*code, template.PH_RL_CONVERT_IN_TO_DETAIL_GRPC, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_ASSIGN_IN_DETAIL_GRPC, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_CREATE_ASSIGN_TO_BO_GRPC, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_CONVERT_IN_TO_LISTINFO_GRPC, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_ASSIGN_IN_LISTINFO_GRPC, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_GRPC_HANDLER_FUNCTIONS, "")
		// 其他表类型不需要BR表支持
		*code = strings.ReplaceAll(*code, template.PH_BR_HTTP_HANDLER_FUNCTIONS, "")
		*code = strings.ReplaceAll(*code, template.PH_BR_GRPC_HANDLER_FUNCTIONS, "")
	}

	genAssignBOToVOGRPC(code, table)
	genAssignBOToVO(code, table)
	genColListInVO(code, table)
	genColListForCreate(code, table)
	genAssignCreateToBO(code, table)
	genAssignCreateToBOGRPC(code, table)
	genColListForFilter(code, table)
	getColListForOrder(code, table)
	genColListForFilterDoc(code, table)
	genColListForOrderDoc(code, table)
	genAssignFilterToOption(code, table)
	genAssignFilterToOptionGRPC(code, table)
	genAssignOrderToOption(code, table)
	genColListForList(code, table)
	genAssignBOForList(code, table)
	genColListToSelectForList(code, table)
	genColListForUpdate(code, table)
	genAssignUpdateToSet(code, table)
	genAssignUpdateToSetGRPC(code, table)

	project := table.Database.Project
	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, project.GoModule)
	*code = strings.ReplaceAll(*code, template.PH_GO_VERSION, project.GoVersion)
	*code = strings.ReplaceAll(*code, template.PH_ENV_PREFIX, helper.GetEnvPrefix(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME, project.Name)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_DIR, helper.GetDirName(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_PKG, helper.GetPkgName(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_STRUCT, helper.GetStructName(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_TABLE_COMMENT, table.Comment)
	*code = strings.ReplaceAll(*code, template.PH_TABLE_NAME_STRUCT, helper.GetStructName(table.Name))
	*code = strings.ReplaceAll(*code, template.PH_TABLE_NAME_URI, helper.GetURIName(table.Name))
	*code = strings.ReplaceAll(*code, template.PH_TABLE_NAME, table.Name)
	*code = strings.ReplaceAll(*code, template.PH_TABLE_NAME_VAR, helper.GetVarName(table.Name))
}

func genColListInVO(code *string, table *model.Table) {
	var (
		buf    strings.Builder
		goType string
		err    error
	)
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}

		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {
			goType = "string"
			if !col.IsRequired {
				goType = "*" + goType
			}
		} else {
			goType, err = helper.GetGoType(col)
			if err != nil {
				log.Fatalf("fail to get to type: %v", err)
			}
			if !col.IsRequired && !helper.IsGoTypeNullable(goType) {
				goType = "*" + goType
			}
		}
		comment := col.Comment
		if !col.IsRequired {
			comment += " (nullable)"
		}
		buf.WriteString(fmt.Sprintf("\n%s %s `json:\"%s\"` // %s",
			helper.GetStructName(col.Name),
			goType,
			helper.GetDirName(col.Name),
			comment))
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_IN_VO, buf.String())
}

func genAssignModelToVO(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}

		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {

			tFormat := "SecondTimeFormat"
			if col.Type == model.ColumnType_TIME ||
				col.Type == model.ColumnType_TIMETZ {
				tFormat = "HourMinuteSecondFormat"
			} else if col.Type == model.ColumnType_DATE {
				tFormat = "DateFormat"
			}

			if !col.IsRequired {
				bufPrepare.WriteString(fmt.Sprintf("var %s *string\n",
					helper.GetVarName(col.Name)))
				bufPrepare.WriteString(fmt.Sprintf("if m%s.%s != nil {\n",
					helper.GetStructName(table.Name),
					helper.GetTableColName(col.Name),
				))
				bufPrepare.WriteString(fmt.Sprintf("*%s = m%s.%s.Format(constraint.%s)\n",
					helper.GetVarName(col.Name), helper.GetStructName(table.Name),
					helper.GetTableColName(col.Name),
					tFormat,
				))
				bufPrepare.WriteString("}\n")
				buf.WriteString(fmt.Sprintf("\n%s: %s,",
					helper.GetStructName(col.Name),
					helper.GetVarName(col.Name)))
			} else {
				buf.WriteString(fmt.Sprintf("\n%s: m%s.%s.Format(constraint.%s),",
					helper.GetStructName(col.Name),
					helper.GetStructName(table.Name),
					helper.GetTableColName(col.Name),
					tFormat,
				))
			}
		} else {
			buf.WriteString(fmt.Sprintf("\n%s: m%s.%s,",
				helper.GetStructName(col.Name),
				helper.GetStructName(table.Name),
				helper.GetTableColName(col.Name)))
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_MODEL_TO_VO, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_MODEL_TO_VO, buf.String())
}

func genAssignBOToVO(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}

		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {

			tFormat := "SecondTimeFormat"
			if col.Type == model.ColumnType_TIME ||
				col.Type == model.ColumnType_TIMETZ {
				tFormat = "HourMinuteSecondFormat"
			} else if col.Type == model.ColumnType_DATE {
				tFormat = "DateFormat"
			}

			if !col.IsRequired {
				bufPrepare.WriteString(fmt.Sprintf("var %s *string\n",
					helper.GetVarName(col.Name)))
				bufPrepare.WriteString(fmt.Sprintf("if bo.%s != nil {\n",
					helper.GetTableColName(col.Name),
				))
				bufPrepare.WriteString(fmt.Sprintf("*%s = bo.%s.Format(constraint.%s)\n",
					helper.GetVarName(col.Name),
					helper.GetTableColName(col.Name),
					tFormat,
				))
				bufPrepare.WriteString("}\n")
				buf.WriteString(fmt.Sprintf("\n%s: %s,",
					helper.GetStructName(col.Name),
					helper.GetVarName(col.Name)))
			} else {
				buf.WriteString(fmt.Sprintf("\n%s: bo.%s.Format(constraint.%s),",
					helper.GetStructName(col.Name),
					helper.GetTableColName(col.Name),
					tFormat,
				))
			}
		} else {
			buf.WriteString(fmt.Sprintf("\n%s: bo.%s,",
				helper.GetStructName(col.Name),
				helper.GetTableColName(col.Name)))
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_BO_TO_VO, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_BO_TO_VO, buf.String())
}

func genAssignModelToVOGRPC(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}

		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {

			tFormat := "SecondTimeFormat"
			if col.Type == model.ColumnType_TIME ||
				col.Type == model.ColumnType_TIMETZ {
				tFormat = "HourMinuteSecondFormat"
			} else if col.Type == model.ColumnType_DATE {
				tFormat = "DateFormat"
			}

			if !col.IsRequired {
				bufPrepare.WriteString(fmt.Sprintf("var %s *string\n",
					helper.GetVarName(col.Name)))
				bufPrepare.WriteString(fmt.Sprintf("if m%s.%s != nil {\n",
					helper.GetStructName(table.Name),
					helper.GetTableColName(col.Name),
				))
				bufPrepare.WriteString(fmt.Sprintf("*%s = m%s.%s.Format(constraint.%s)\n",
					helper.GetVarName(col.Name), helper.GetStructName(table.Name),
					helper.GetTableColName(col.Name),
					tFormat,
				))
				bufPrepare.WriteString("}\n")
				buf.WriteString(fmt.Sprintf("\n%s: %s,",
					helper.GetStructName(col.Name),
					helper.GetVarName(col.Name)))
			} else {
				buf.WriteString(fmt.Sprintf("\n%s: m%s.%s.Format(constraint.%s),",
					helper.GetStructName(col.Name),
					helper.GetStructName(table.Name),
					helper.GetTableColName(col.Name),
					tFormat,
				))
			}
		} else {
			pt, _ := helper.GetProto3ValueType(col)
			pt, _ = helper.Proto3ValueType2GoType(pt)
			gt, _ := helper.GetGoType(col)
			if gt != pt {
				if !col.IsRequired && !helper.IsGoTypeNullable(pt) {
					bufPrepare.WriteString(fmt.Sprintf("var %s *%s\n",
						helper.GetVarName(col.Name), pt))
					bufPrepare.WriteString(fmt.Sprintf("if m%s.%s != nil {\n",
						helper.GetStructName(table.Name),
						helper.GetTableColName(col.Name),
					))
					bufPrepare.WriteString(fmt.Sprintf("*%s = %s(*m%s.%s)\n",
						helper.GetVarName(col.Name),
						pt,
						helper.GetStructName(table.Name),
						helper.GetTableColName(col.Name),
					))
					bufPrepare.WriteString("}\n")
					buf.WriteString(fmt.Sprintf("\n%s: %s,",
						helper.GetStructName(col.Name),
						helper.GetVarName(col.Name)))
				} else {
					buf.WriteString(fmt.Sprintf("\n%s: %s(m%s.%s),",
						helper.GetStructName(col.Name),
						pt,
						helper.GetStructName(table.Name),
						helper.GetTableColName(col.Name)))
				}
			} else {
				buf.WriteString(fmt.Sprintf("\n%s: m%s.%s,",
					helper.GetStructName(col.Name),
					helper.GetStructName(table.Name),
					helper.GetTableColName(col.Name)))
			}
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_MODEL_TO_VO, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_MODEL_TO_VO_GRPC, buf.String())
}

func genAssignBOToVOGRPC(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}

		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {

			tFormat := "SecondTimeFormat"
			if col.Type == model.ColumnType_TIME ||
				col.Type == model.ColumnType_TIMETZ {
				tFormat = "HourMinuteSecondFormat"
			} else if col.Type == model.ColumnType_DATE {
				tFormat = "DateFormat"
			}

			if !col.IsRequired {
				bufPrepare.WriteString(fmt.Sprintf("var %s *string\n",
					helper.GetVarName(col.Name)))
				bufPrepare.WriteString(fmt.Sprintf("if bo.%s != nil {\n",
					helper.GetTableColName(col.Name),
				))
				bufPrepare.WriteString(fmt.Sprintf("*%s = bo.%s.Format(constraint.%s)\n",
					helper.GetVarName(col.Name),
					helper.GetTableColName(col.Name),
					tFormat,
				))
				bufPrepare.WriteString("}\n")
				buf.WriteString(fmt.Sprintf("\n%s: %s,",
					helper.GetStructName(col.Name),
					helper.GetVarName(col.Name)))
			} else {
				buf.WriteString(fmt.Sprintf("\n%s: bo.%s.Format(constraint.%s),",
					helper.GetStructName(col.Name),
					helper.GetTableColName(col.Name),
					tFormat,
				))
			}
		} else {
			pt, _ := helper.GetProto3ValueType(col)
			pt, _ = helper.Proto3ValueType2GoType(pt)
			gt, _ := helper.GetGoType(col)
			if gt != pt {
				if !col.IsRequired && !helper.IsGoTypeNullable(pt) {
					bufPrepare.WriteString(fmt.Sprintf("var %s *%s\n",
						helper.GetVarName(col.Name), pt))
					bufPrepare.WriteString(fmt.Sprintf("if m%s.%s != nil {\n",
						helper.GetStructName(table.Name),
						helper.GetTableColName(col.Name),
					))
					bufPrepare.WriteString(fmt.Sprintf("*%s = %s(*m%s.%s)\n",
						helper.GetVarName(col.Name),
						pt,
						helper.GetStructName(table.Name),
						helper.GetTableColName(col.Name),
					))
					bufPrepare.WriteString("}\n")
					buf.WriteString(fmt.Sprintf("\n%s: %s,",
						helper.GetStructName(col.Name),
						helper.GetVarName(col.Name)))
				} else {
					buf.WriteString(fmt.Sprintf("\n%s: %s(bo.%s),",
						helper.GetStructName(col.Name),
						pt,
						helper.GetTableColName(col.Name)))
				}
			} else {
				buf.WriteString(fmt.Sprintf("\n%s: bo.%s,",
					helper.GetStructName(col.Name),
					helper.GetTableColName(col.Name)))
			}
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_BO_TO_VO_GRPC, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_BO_TO_VO_GRPC, buf.String())
}

func generateGoFile(path string, tpl string, project *model.Project, addHeader func(s string) string) error {
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
	replaceTpl(&code, project)

	// go format
	formatted, err := jgstr.FormatGo([]byte(code))
	if err != nil {
		log.Errorf("format source failed: %v\n%s", err, code)
		return err
	}
	_, _ = f.Write(formatted)
	return nil
}

func replaceTpl(code *string, project *model.Project) {
	genUsecaseListInStruct(code, project)
	genUsecaseListInArg(code, project)
	genAssignUsecaseList(code, project)
	genImportUsecaseList(code, project)

	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, project.GoModule)
	*code = strings.ReplaceAll(*code, template.PH_GO_VERSION, project.GoVersion)
	*code = strings.ReplaceAll(*code, template.PH_ENV_PREFIX, helper.GetEnvPrefix(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME, project.Name)
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_DIR, helper.GetDirName(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_PKG, helper.GetPkgName(project.Name))
	*code = strings.ReplaceAll(*code, template.PH_PROJECT_NAME_STRUCT, helper.GetStructName(project.Name))

	if project.Database == nil || len(project.Database.Tables) == 0 {
		*code = strings.ReplaceAll(*code, template.PH_COMMENT_REPO, "//")
		if len(project.Domain.Usecases) == 0 {
			*code = strings.ReplaceAll(*code, template.PH_COMMENT_DOMAIN, "//")
		} else {
			*code = strings.ReplaceAll(*code, template.PH_COMMENT_DOMAIN, "")
		}
	} else {
		*code = strings.ReplaceAll(*code, template.PH_COMMENT_REPO, "")
		if len(project.Domain.Usecases) == 0 {
			*code = strings.ReplaceAll(*code, template.PH_COMMENT_DOMAIN, "//")
		} else {
			*code = strings.ReplaceAll(*code, template.PH_COMMENT_DOMAIN, "")
		}
	}
}

func genUsecaseListInStruct(code *string, project *model.Project) {
	// usecase list
	var buf strings.Builder
	for _, usecase := range project.Domain.Usecases {
		buf.WriteString(fmt.Sprintf("%sUsecase *%s.%sUsecase\n",
			helper.GetStructName(usecase.Name),
			helper.GetPkgName(usecase.Name),
			helper.GetStructName(usecase.Name)),
		)
	}
	*code = strings.ReplaceAll(*code, template.PH_USECASE_LIST_IN_STRUCT, buf.String())
}

func genUsecaseListInArg(code *string, project *model.Project) {
	// usecase list
	var buf strings.Builder
	for _, usecase := range project.Domain.Usecases {
		buf.WriteString(fmt.Sprintf("%sUsecase *%s.%sUsecase,\n",
			helper.GetVarName(usecase.Name),
			helper.GetPkgName(usecase.Name),
			helper.GetStructName(usecase.Name)),
		)
	}
	*code = strings.ReplaceAll(*code, template.PH_USECASE_LIST_IN_ARG, buf.String())
}

func genAssignUsecaseList(code *string, project *model.Project) {
	if len(project.Domain.Usecases) == 0 {
		*code = strings.ReplaceAll(*code, template.PH_ASSIGN_USECASE_LIST, "")
		return
	}
	// usecase list
	var buf strings.Builder
	for _, usecase := range project.Domain.Usecases {
		buf.WriteString(fmt.Sprintf("%sUsecase: %sUsecase,\n",
			helper.GetStructName(usecase.Name),
			helper.GetVarName(usecase.Name)),
		)
	}
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_USECASE_LIST, buf.String())
}

func genImportUsecaseList(code *string, project *model.Project) {
	// import usecase list
	var buf strings.Builder
	for _, usecase := range project.Domain.Usecases {
		buf.WriteString(fmt.Sprintf("%s \"%s/internal/usecase/%s\"\n",
			helper.GetPkgName(usecase.Name),
			project.GoModule,
			helper.GetDirName(usecase.Name)),
		)
	}
	*code = strings.ReplaceAll(*code, template.PH_IMPORT_USECASE_LIST, buf.String())
}

// PH_COL_LIST_FOR_CREATE
// like:
//
//	Name   *string `json:"name" binding:"omitempty,min=1,max=64"` // 用户名
//	Gender uint64  `json:"gender" binding:"required,gte=1"`       // 性别
func genColListForCreate(code *string, table *model.Table) {
	var buf strings.Builder
	for _, col := range table.Columns {
		if !col.IsAlterable {
			continue
		}
		if col.IsHidden {
			continue
		}
		buf.WriteString(fmt.Sprintf("\n%s %s `json:\"%s\"%s` // %s",
			helper.GetStructName(col.Name),
			helper.GetGoTypeForHandler(col),
			col.Name,
			helper.GetBinding(col),
			helper.GetCommentForHandler(col),
		))
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_CREATE, buf.String())
}

// PH_ASSIGN_CREATE_TO_MODEL
// like:
//
//	Name:   req.Name,
//	Gender: req.Gender,
func genAssignCreateToModel(code *string, table *model.Table) {
	var buf strings.Builder
	for _, col := range table.Columns {
		if !col.IsAlterable {
			continue
		}
		if col.IsHidden {
			continue
		}
		buf.WriteString(fmt.Sprintf("%s: req.%s,\n",
			helper.GetTableColName(col.Name),
			helper.GetStructName(col.Name),
		))
	}
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_CREATE_TO_MODEL, buf.String())
}

// PH_ASSIGN_CREATE_TO_BO
// like:
//
//	Name:   req.Name,
//	Gender: req.Gender,
func genAssignCreateToBO(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	for _, col := range table.Columns {
		if !col.IsAlterable {
			continue
		}
		if col.IsHidden {
			continue
		}
		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {

			tFormat := "SecondTimeFormat"
			if col.Type == model.ColumnType_TIME ||
				col.Type == model.ColumnType_TIMETZ {
				tFormat = "HourMinuteSecondFormat"
			} else if col.Type == model.ColumnType_DATE {
				tFormat = "DateFormat"
			}

			if !col.IsRequired {
				bufPrepare.WriteString(fmt.Sprintf("var %s *time.Time\n",
					helper.GetVarName(col.Name)))
				bufPrepare.WriteString(fmt.Sprintf("if req.%s != nil {\n",
					helper.GetStructName(col.Name),
				))
				bufPrepare.WriteString(fmt.Sprintf("t, err := time.ParseInLocation(constraint.%s, *req.%s, time.Local)\n",
					tFormat,
					helper.GetStructName(col.Name),
				))
				bufPrepare.WriteString("if err != nil {\n")
				bufPrepare.WriteString(`log.WithError(err).Errorf("fail to parse time: %s", *req.` +
					helper.GetStructName(col.Name) +
					")\n")
				bufPrepare.WriteString("ResponseFail(c, cerror.InvalidArgument(\"" + helper.GetStructName(col.Name) + "\"))\n")
				bufPrepare.WriteString("return\n")
				bufPrepare.WriteString("}\n")
				bufPrepare.WriteString(fmt.Sprintf("%s = &t\n", helper.GetVarName(col.Name)))
				bufPrepare.WriteString("}\n")
				buf.WriteString(fmt.Sprintf("\n%s: %s,",
					helper.GetTableColName(col.Name),
					helper.GetVarName(col.Name)))
			} else {
				bufPrepare.WriteString(fmt.Sprintf("%s, err := time.ParseInLocation(constraint.%s, req.%s, time.Local)\n",
					helper.GetVarName(col.Name),
					tFormat,
					helper.GetStructName(col.Name),
				))
				bufPrepare.WriteString("if err != nil {\n")
				bufPrepare.WriteString(`log.WithError(err).Errorf("fail to parse time: %s", req.` +
					helper.GetStructName(col.Name) +
					")\n")
				bufPrepare.WriteString("ResponseFail(c, cerror.InvalidArgument(\"" + helper.GetStructName(col.Name) + "\"))\n")
				bufPrepare.WriteString("return\n")
				bufPrepare.WriteString("}\n")
				buf.WriteString(fmt.Sprintf("\n%s: %s,",
					helper.GetTableColName(col.Name),
					helper.GetVarName(col.Name)))
			}
		} else {
			buf.WriteString(fmt.Sprintf("\n%s: req.%s,",
				helper.GetTableColName(col.Name),
				helper.GetStructName(col.Name),
			))
		}
	}

	// 为DATA表添加RL字段的赋值
	if table.Type == model.TableType_DATA {
		rlTables := helper.GetMainTableRLs(table, table.Database.Tables)
		for _, rlTable := range rlTables {
			rlVarName := helper.GetVarName(rlTable.Name)
			rlStructName := helper.GetStructName(rlTable.Name)
			buf.WriteString(fmt.Sprintf("\n%ss: %ss,", rlStructName, rlVarName))
		}
	}

	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_CREATE_TO_BO, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_CREATE_TO_BO, buf.String())
}

// PH_ASSIGN_CREATE_TO_BO_GRPC
// like:
//
//	Name:   req.Name,
//	Gender: req.Gender,
func genAssignCreateToBOGRPC(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	for _, col := range table.Columns {
		if !col.IsAlterable {
			continue
		}
		if col.IsHidden {
			continue
		}
		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {

			tFormat := "SecondTimeFormat"
			if col.Type == model.ColumnType_TIME ||
				col.Type == model.ColumnType_TIMETZ {
				tFormat = "HourMinuteSecondFormat"
			} else if col.Type == model.ColumnType_DATE {
				tFormat = "DateFormat"
			}

			if !col.IsRequired {
				bufPrepare.WriteString(fmt.Sprintf("var %s *time.Time\n",
					helper.GetVarName(col.Name)))
				bufPrepare.WriteString(fmt.Sprintf("if req.%s != nil {\n",
					helper.GetStructName(col.Name),
				))
				bufPrepare.WriteString(fmt.Sprintf("t, err := time.ParseInLocation(constraint.%s, *req.%s, time.Local)\n",
					tFormat,
					helper.GetStructName(col.Name),
				))
				bufPrepare.WriteString("if err != nil {\n")
				bufPrepare.WriteString(`log.WithError(err).Errorf("fail to parse time: %s", *req.` +
					helper.GetStructName(col.Name) +
					")\n")
				bufPrepare.WriteString("return nil, cerror.InvalidArgument(\"" + helper.GetStructName(col.Name) + "\")\n")
				bufPrepare.WriteString("}\n")
				bufPrepare.WriteString(fmt.Sprintf("%s = &t\n", helper.GetVarName(col.Name)))
				bufPrepare.WriteString("}\n")
				buf.WriteString(fmt.Sprintf("\n%s: %s,",
					helper.GetTableColName(col.Name),
					helper.GetVarName(col.Name)))
			} else {
				bufPrepare.WriteString(fmt.Sprintf("%s, err := time.ParseInLocation(constraint.%s, req.%s, time.Local)\n",
					helper.GetVarName(col.Name),
					tFormat,
					helper.GetStructName(col.Name),
				))
				bufPrepare.WriteString("if err != nil {\n")
				bufPrepare.WriteString(`log.WithError(err).Errorf("fail to parse time: %s", req.` +
					helper.GetStructName(col.Name) +
					")\n")
				bufPrepare.WriteString("return nil, cerror.InvalidArgument(\"" + helper.GetStructName(col.Name) + "\")\n")
				bufPrepare.WriteString("}\n")
				buf.WriteString(fmt.Sprintf("\n%s: %s,",
					helper.GetTableColName(col.Name),
					helper.GetVarName(col.Name)))
			}
		} else {
			pt, _ := helper.GetProto3ValueType(col)
			pt, _ = helper.Proto3ValueType2GoType(pt)
			gt, _ := helper.GetGoType(col)
			if gt != pt {
				if !col.IsRequired && !helper.IsGoTypeNullable(gt) {
					bufPrepare.WriteString(fmt.Sprintf("var %s *%s\n",
						helper.GetVarName(col.Name), gt))
					bufPrepare.WriteString(fmt.Sprintf("if req.%s != nil {\n",
						helper.GetStructName(col.Name),
					))
					bufPrepare.WriteString(fmt.Sprintf("*%s = %s(*req.%s)\n",
						helper.GetVarName(col.Name),
						gt,
						helper.GetStructName(col.Name),
					))
					bufPrepare.WriteString("}\n")
					buf.WriteString(fmt.Sprintf("\n%s: %s,",
						helper.GetTableColName(col.Name),
						helper.GetVarName(col.Name)))
				} else {
					buf.WriteString(fmt.Sprintf("\n%s: %s(req.%s),",
						helper.GetTableColName(col.Name),
						gt,
						helper.GetStructName(col.Name),
					))
				}
			} else {
				buf.WriteString(fmt.Sprintf("\n%s: req.%s,",
					helper.GetTableColName(col.Name),
					helper.GetStructName(col.Name),
				))
			}
		}
	}

	// 为DATA表添加RL字段的赋值
	if table.Type == model.TableType_DATA {
		rlTables := helper.GetMainTableRLs(table, table.Database.Tables)
		for _, rlTable := range rlTables {
			rlVarName := helper.GetVarName(rlTable.Name)
			rlStructName := helper.GetStructName(rlTable.Name)
			buf.WriteString(fmt.Sprintf("\n%ss: %ss,", rlStructName, rlVarName))
		}
	}

	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_CREATE_TO_BO_GRPC, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_CREATE_TO_BO_GRPC, buf.String())
}

// PH_COL_LIST_FOR_FILTER
// like:
// // 筛选条件
// Gender *uint64 `form:"gender" binding:"omitempty,gte=1"` // 性别
func genColListForFilter(code *string, table *model.Table) {
	var buf strings.Builder
	hasFilterCol := false
	for _, col := range table.Columns {
		if col.IsFilter && !col.IsHidden {
			hasFilterCol = true
			break
		}
	}
	if !hasFilterCol {
		*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_FILTER, "")
		return
	}
	buf.WriteString("// 筛选条件\n")
	for _, col := range table.Columns {
		if !col.IsFilter || col.IsHidden {
			continue
		}
		gotype := helper.GetGoTypeForHandler(col)
		if !strings.HasPrefix(gotype, "*") && !helper.IsGoTypeNullable(gotype) {
			gotype = "*" + gotype
		}
		binding := helper.GetBinding(col)
		if binding != "" {
			binding = strings.ReplaceAll(binding, "required", "omitempty")
		} else {
			binding = " binding:\"omitempty\""
		}
		buf.WriteString(fmt.Sprintf("%s %s `form:\"%s\"%s` // %s\n",
			helper.GetStructName(col.Name),
			gotype,
			col.Name,
			binding,
			helper.GetCommentForHandler(col),
		))
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_FILTER, buf.String())
}

// PH_COL_LIST_FOR_ORDER
// like:
// // 排序条件
// OrderBy   *string           `form:"order_by" binding:"omitempty,oneof=id"`         // 排序字段,默认id
// OrderType *option.OrderType `form:"order_type" binding:"omitempty,oneof=asc desc"` // 排序类型,默认desc
func getColListForOrder(code *string, table *model.Table) {
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
		return
	}
	buf.WriteString("// 排序条件\n")
	orderCols := make([]string, 0)
	for _, col := range table.Columns {
		if !col.IsOrder || col.IsHidden {
			continue
		}
		orderCols = append(orderCols, col.Name)
	}
	buf.WriteString(fmt.Sprintf("OrderBy *string `form:\"order_by\" binding:\"omitempty,oneof=%s\"` // 排序字段,可选:%s\n",
		strings.Join(orderCols, " "),
		strings.Join(orderCols, "|"),
	))
	buf.WriteString("OrderType *string `form:\"order_type\" binding:\"omitempty,oneof=asc desc\"` // 排序类型,默认desc\n")
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_ORDER, buf.String())
}

// PH_COL_LIST_FOR_FILTER_DOC
// like:
// // @Param			gender			query		int		false	"性别"
func genColListForFilterDoc(code *string, table *model.Table) {
	var buf strings.Builder
	hasFilterCol := false
	for _, col := range table.Columns {
		if col.IsFilter && !col.IsHidden {
			hasFilterCol = true
			break
		}
	}
	if !hasFilterCol {
		*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_FILTER_DOC, "")
		return
	}
	for _, col := range table.Columns {
		if !col.IsFilter || col.IsHidden {
			continue
		}
		gotype := helper.GetGoTypeForHandler(col)
		gotype = strings.TrimPrefix(gotype, "*")
		buf.WriteString(fmt.Sprintf("\n//	@Param		%s			query		%s	false	\"%s\"",
			col.Name,
			gotype,
			helper.GetCommentForHandler(col),
		))
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_FILTER_DOC, buf.String())
}

// PH_COL_LIST_FOR_ORDER_DOC
// like:
// // @Param order_by query string false "排序字段,默认id"
// // @Param order_type query string false "排序类型,默认desc"
func genColListForOrderDoc(code *string, table *model.Table) {
	var buf strings.Builder
	hasOrderCol := false
	for _, col := range table.Columns {
		if col.IsOrder && !col.IsHidden {
			hasOrderCol = true
			break
		}
	}
	if !hasOrderCol {
		*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_ORDER_DOC, "")
		return
	}
	orderCols := make([]string, 0)
	for _, col := range table.Columns {
		if !col.IsOrder || col.IsHidden {
			continue
		}
		orderCols = append(orderCols, col.Name)
	}
	buf.WriteString(fmt.Sprintf("\n//	@Param		order_by		query		string	false	\"排序字段, 可选: %s\"",
		strings.Join(orderCols, "|"),
	))
	buf.WriteString("\n//	@Param		order_type		query		string	false	\"排序类型,默认desc\"")
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_ORDER_DOC, buf.String())
}

// PH_ASSIGN_FILTER_TO_OPTION
// like:
//
//	Filter: &domain.UserFilterOption{
//		Gender: req.Gender,
//	},
func genAssignFilterToOption(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	hasFilterCol := false
	for _, col := range table.Columns {
		if col.IsFilter && !col.IsHidden {
			hasFilterCol = true
			break
		}
	}
	if !hasFilterCol {
		*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_FILTER_TO_OPTION, "")
		*code = strings.ReplaceAll(*code, template.PH_ASSIGN_FILTER_TO_OPTION, "")
		return
	}
	buf.WriteString(fmt.Sprintf("\nFilter: &biz.%sFilterOption{", helper.GetStructName(table.Name)))
	for _, col := range table.Columns {
		if !col.IsFilter || col.IsHidden {
			continue
		}
		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {

			tFormat := "SecondTimeFormat"
			if col.Type == model.ColumnType_TIME ||
				col.Type == model.ColumnType_TIMETZ {
				tFormat = "HourMinuteSecondFormat"
			} else if col.Type == model.ColumnType_DATE {
				tFormat = "DateFormat"
			}

			bufPrepare.WriteString(fmt.Sprintf("var %s *time.Time\n",
				helper.GetVarName(col.Name)))
			bufPrepare.WriteString(fmt.Sprintf("if req.%s != nil {\n",
				helper.GetStructName(col.Name),
			))
			bufPrepare.WriteString(fmt.Sprintf("t, err := time.ParseInLocation(constraint.%s, *req.%s, time.Local)\n",
				tFormat,
				helper.GetStructName(col.Name),
			))
			bufPrepare.WriteString("if err != nil {\n")
			bufPrepare.WriteString(`log.WithError(err).Errorf("fail to parse time: %s", *req.` +
				helper.GetStructName(col.Name) +
				")\n")
			bufPrepare.WriteString("ResponseFail(c, cerror.InvalidArgument(\"" + helper.GetStructName(col.Name) + "\"))\n")
			bufPrepare.WriteString("return\n")
			bufPrepare.WriteString("}\n")
			bufPrepare.WriteString(fmt.Sprintf("%s = &t\n", helper.GetVarName(col.Name)))
			bufPrepare.WriteString("}\n")
			buf.WriteString(fmt.Sprintf("\n%s: %s,",
				helper.GetTableColName(col.Name),
				helper.GetVarName(col.Name)))
		} else {
			buf.WriteString(fmt.Sprintf("\n%s: req.%s,",
				helper.GetTableColName(col.Name),
				helper.GetStructName(col.Name),
			))
		}
	}
	buf.WriteString("\n},")
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_FILTER_TO_OPTION, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_FILTER_TO_OPTION, buf.String())
}

// PH_ASSIGN_FILTER_TO_OPTION_GRPC
// like:
//
//	Filter: &domain.UserFilterOption{
//		Gender: req.Gender,
//	},
func genAssignFilterToOptionGRPC(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	hasFilterCol := false
	for _, col := range table.Columns {
		if col.IsFilter && !col.IsHidden {
			hasFilterCol = true
			break
		}
	}
	if !hasFilterCol {
		*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_FILTER_TO_OPTION_GRPC, "")
		*code = strings.ReplaceAll(*code, template.PH_ASSIGN_FILTER_TO_OPTION_GRPC, "")
		return
	}
	buf.WriteString(fmt.Sprintf("\nFilter: &biz.%sFilterOption{", helper.GetStructName(table.Name)))
	for _, col := range table.Columns {
		if !col.IsFilter || col.IsHidden {
			continue
		}
		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {

			tFormat := "SecondTimeFormat"
			if col.Type == model.ColumnType_TIME ||
				col.Type == model.ColumnType_TIMETZ {
				tFormat = "HourMinuteSecondFormat"
			} else if col.Type == model.ColumnType_DATE {
				tFormat = "DateFormat"
			}

			bufPrepare.WriteString(fmt.Sprintf("var %s *time.Time\n",
				helper.GetVarName(col.Name)))
			bufPrepare.WriteString(fmt.Sprintf("if req.%s != nil {\n",
				helper.GetStructName(col.Name),
			))
			bufPrepare.WriteString(fmt.Sprintf("t, err := time.ParseInLocation(constraint.%s, *req.%s, time.Local)\n",
				tFormat,
				helper.GetStructName(col.Name),
			))
			bufPrepare.WriteString("if err != nil {\n")
			bufPrepare.WriteString(`log.WithError(err).Errorf("fail to parse time: %s", *req.` +
				helper.GetStructName(col.Name) +
				")\n")
			bufPrepare.WriteString("return nil, cerror.InvalidArgument(\"" + helper.GetStructName(col.Name) + "\")\n")
			bufPrepare.WriteString("}\n")
			bufPrepare.WriteString(fmt.Sprintf("%s = &t\n", helper.GetVarName(col.Name)))
			bufPrepare.WriteString("}\n")
			buf.WriteString(fmt.Sprintf("\n%s: %s,",
				helper.GetTableColName(col.Name),
				helper.GetVarName(col.Name)))
		} else {
			pt, _ := helper.GetProto3ValueType(col)
			pt, _ = helper.Proto3ValueType2GoType(pt)
			gt, _ := helper.GetGoType(col)
			if gt != pt {
				if !helper.IsGoTypeNullable(gt) {
					bufPrepare.WriteString(fmt.Sprintf("var %s *%s\n",
						helper.GetVarName(col.Name), gt))
					bufPrepare.WriteString(fmt.Sprintf("if req.%s != nil {\n",
						helper.GetStructName(col.Name),
					))
					bufPrepare.WriteString(fmt.Sprintf("*%s = %s(*req.%s)\n",
						helper.GetVarName(col.Name),
						gt,
						helper.GetStructName(col.Name),
					))
					bufPrepare.WriteString("}\n")
					buf.WriteString(fmt.Sprintf("\n%s: %s,",
						helper.GetTableColName(col.Name),
						helper.GetVarName(col.Name)))
				} else {
					buf.WriteString(fmt.Sprintf("\n%s: %s(req.%s),",
						helper.GetTableColName(col.Name),
						gt,
						helper.GetStructName(col.Name),
					))
				}
			} else {
				buf.WriteString(fmt.Sprintf("\n%s: req.%s,",
					helper.GetTableColName(col.Name),
					helper.GetStructName(col.Name),
				))
			}
		}
	}
	buf.WriteString("\n},")
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_FILTER_TO_OPTION_GRPC, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_FILTER_TO_OPTION_GRPC, buf.String())
}

// PH_ASSIGN_ORDER_TO_OPTION
// like:
//
//	Order: &option.OrderOption{
//		OrderBy:   req.OrderBy,
//		OrderType: req.OrderType,
//	},
func genAssignOrderToOption(code *string, table *model.Table) {
	var buf strings.Builder
	hasOrderCol := false
	for _, col := range table.Columns {
		if col.IsOrder && !col.IsHidden {
			hasOrderCol = true
			break
		}
	}
	if !hasOrderCol {
		*code = strings.ReplaceAll(*code, template.PH_ASSIGN_ORDER_TO_OPTION, "")
		return
	}
	buf.WriteString("\nOrder: &option.OrderOption{")
	buf.WriteString("\nOrderBy:   req.OrderBy,")
	buf.WriteString("\nOrderType: req.OrderType,")
	buf.WriteString("\n},")
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_ORDER_TO_OPTION, buf.String())
}

// PH_COL_LIST_FOR_LIST
// like:
// Id     uint64  `json:"id"`     //
// Name   *string `json:"name"`   // 用户名
// Gender uint64  `json:"gender"` // 性别
func genColListForList(code *string, table *model.Table) {
	var (
		buf    strings.Builder
		goType string
		err    error
	)
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}
		if !col.IsList {
			continue
		}

		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {
			goType = "string"
			if !col.IsRequired {
				goType = "*" + goType
			}
		} else {
			goType, err = helper.GetGoType(col)
			if err != nil {
				log.Fatalf("fail to get to type: %v", err)
			}
			if !col.IsRequired && !helper.IsGoTypeNullable(goType) {
				goType = "*" + goType
			}
		}
		comment := col.Comment
		if !col.IsRequired {
			comment += " (nullable)"
		}
		buf.WriteString(fmt.Sprintf("\n%s %s `json:\"%s\"` // %s",
			helper.GetStructName(col.Name),
			goType,
			helper.GetDirName(col.Name),
			comment))
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_LIST, buf.String())
}

// PH_ASSIGN_MODEL_FOR_LIST
// like:
// Id:     mUser.ID,
// Name:   mUser.Name,
// Gender: mUser.Gender,
func genAssignModelForList(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}
		if !col.IsList {
			continue
		}

		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {

			tFormat := "SecondTimeFormat"
			if col.Type == model.ColumnType_TIME ||
				col.Type == model.ColumnType_TIMETZ {
				tFormat = "HourMinuteSecondFormat"
			} else if col.Type == model.ColumnType_DATE {
				tFormat = "DateFormat"
			}

			if !col.IsRequired {
				bufPrepare.WriteString(fmt.Sprintf("var %s *string\n",
					helper.GetVarName(col.Name)))
				bufPrepare.WriteString(fmt.Sprintf("if ms%s[i].%s != nil {\n",
					helper.GetStructName(table.Name),
					helper.GetTableColName(col.Name),
				))
				bufPrepare.WriteString(fmt.Sprintf("*%s = ms%s[i].%s.Format(constraint.%s)\n",
					helper.GetVarName(col.Name), helper.GetStructName(table.Name),
					helper.GetTableColName(col.Name),
					tFormat,
				))
				bufPrepare.WriteString("}\n")
				buf.WriteString(fmt.Sprintf("\n%s: %s,",
					helper.GetStructName(col.Name),
					helper.GetVarName(col.Name)))
			} else {
				buf.WriteString(fmt.Sprintf("\n%s: ms%s[i].%s.Format(constraint.%s),",
					helper.GetStructName(col.Name),
					helper.GetStructName(table.Name),
					helper.GetTableColName(col.Name),
					tFormat,
				))
			}

		} else {
			buf.WriteString(fmt.Sprintf("\n%s: ms%s[i].%s,",
				helper.GetStructName(col.Name),
				helper.GetStructName(table.Name),
				helper.GetTableColName(col.Name)))
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_MODEL_FOR_LIST, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_MODEL_FOR_LIST, buf.String())
}

// PH_ASSIGN_BO_FOR_LIST
// like:
// Id:     mUser.ID,
// Name:   mUser.Name,
// Gender: mUser.Gender,
func genAssignBOForList(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}
		if !col.IsList {
			continue
		}

		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {

			tFormat := "SecondTimeFormat"
			if col.Type == model.ColumnType_TIME ||
				col.Type == model.ColumnType_TIMETZ {
				tFormat = "HourMinuteSecondFormat"
			} else if col.Type == model.ColumnType_DATE {
				tFormat = "DateFormat"
			}

			if !col.IsRequired {
				bufPrepare.WriteString(fmt.Sprintf("var %s *string\n",
					helper.GetVarName(col.Name)))
				bufPrepare.WriteString(fmt.Sprintf("if objs[i].%s != nil {\n",
					helper.GetTableColName(col.Name),
				))
				bufPrepare.WriteString(fmt.Sprintf("*%s = objs[i].%s.Format(constraint.%s)\n",
					helper.GetVarName(col.Name),
					helper.GetTableColName(col.Name),
					tFormat,
				))
				bufPrepare.WriteString("}\n")
				buf.WriteString(fmt.Sprintf("\n%s: %s,",
					helper.GetStructName(col.Name),
					helper.GetVarName(col.Name)))
			} else {
				buf.WriteString(fmt.Sprintf("\n%s: objs[i].%s.Format(constraint.%s),",
					helper.GetStructName(col.Name),
					helper.GetTableColName(col.Name),
					tFormat,
				))
			}

		} else {
			buf.WriteString(fmt.Sprintf("\n%s: objs[i].%s,",
				helper.GetStructName(col.Name),
				helper.GetTableColName(col.Name)))
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_BO_FOR_LIST, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_BO_FOR_LIST, buf.String())
}

// PH_COL_LIST_TO_SELECT_FOR_LIST
// like:
//
//	model.ColUserID,
//	model.ColUserName,
//	model.ColUserGender,
func genColListToSelectForList(code *string, table *model.Table) {
	var buf strings.Builder
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}
		if !col.IsList {
			continue
		}
		buf.WriteString(fmt.Sprintf("\nmodel.Col%s%s,",
			helper.GetStructName(table.Name),
			helper.GetTableColName(col.Name),
		))
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_TO_SELECT_FOR_LIST, buf.String())
}

// PH_COL_LIST_FOR_UPDATE
// like:
// Name   *string `json:"name" binding:"omitempty,max=64"`  // 用户名
// Gender *uint64 `json:"gender" binding:"omitempty,gte=1"` // 性别
func genColListForUpdate(code *string, table *model.Table) {
	var buf strings.Builder
	for _, col := range table.Columns {
		if !col.IsAlterable {
			continue
		}
		if col.IsHidden {
			continue
		}
		binding := helper.GetBinding(col)
		if binding != "" {
			binding = strings.ReplaceAll(binding, "required", "omitempty")
		} else {
			binding = " binding:\"omitempty\""
		}
		gotype := helper.GetGoTypeForHandler(col)
		if !strings.HasPrefix(gotype, "*") && !helper.IsGoTypeNullable(gotype) {
			gotype = "*" + gotype
		}
		buf.WriteString(fmt.Sprintf("\n%s %s `json:\"%s\"%s` // %s",
			helper.GetStructName(col.Name),
			gotype,
			col.Name,
			binding,
			helper.GetCommentForHandler(col),
		))
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_UPDATE, buf.String())
}

// PH_ASSIGN_UPDATE_TO_SET
// like:
// Name:   req.Name,
// Gender: req.Gender,
func genAssignUpdateToSet(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	for _, col := range table.Columns {
		if !col.IsAlterable {
			continue
		}
		if col.IsHidden {
			continue
		}
		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {

			tFormat := "SecondTimeFormat"
			if col.Type == model.ColumnType_TIME ||
				col.Type == model.ColumnType_TIMETZ {
				tFormat = "HourMinuteSecondFormat"
			} else if col.Type == model.ColumnType_DATE {
				tFormat = "DateFormat"
			}

			bufPrepare.WriteString(fmt.Sprintf("var %s *time.Time\n",
				helper.GetVarName(col.Name)))
			bufPrepare.WriteString(fmt.Sprintf("if req.%s != nil {\n",
				helper.GetStructName(col.Name),
			))
			bufPrepare.WriteString(fmt.Sprintf("t, err := time.ParseInLocation(constraint.%s, *req.%s, time.Local)\n",
				tFormat,
				helper.GetStructName(col.Name),
			))
			bufPrepare.WriteString("if err != nil {\n")
			bufPrepare.WriteString(`log.WithError(err).Errorf("fail to parse time: %s", *req.` +
				helper.GetStructName(col.Name) +
				")\n")
			bufPrepare.WriteString("ResponseFail(c, cerror.InvalidArgument(\"" + helper.GetStructName(col.Name) + "\"))\n")
			bufPrepare.WriteString("return\n")
			bufPrepare.WriteString("}\n")
			bufPrepare.WriteString(fmt.Sprintf("%s = &t\n", helper.GetVarName(col.Name)))
			bufPrepare.WriteString("}\n")
			buf.WriteString(fmt.Sprintf("\n%s: %s,",
				helper.GetTableColName(col.Name),
				helper.GetVarName(col.Name)))
		} else {
			buf.WriteString(fmt.Sprintf("\n%s: req.%s,",
				helper.GetTableColName(col.Name),
				helper.GetStructName(col.Name),
			))
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_UPDATE_TO_SET, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_UPDATE_TO_SET, buf.String())
}

// PH_ASSIGN_UPDATE_TO_SET_GRPC
// like:
// Name:   req.Name,
// Gender: req.Gender,
func genAssignUpdateToSetGRPC(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	for _, col := range table.Columns {
		if !col.IsAlterable {
			continue
		}
		if col.IsHidden {
			continue
		}
		if col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_TIMESTAMP ||
			col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATE ||
			col.Type == model.ColumnType_TIMETZ ||
			col.Type == model.ColumnType_TIMESTAMPTZ {

			tFormat := "SecondTimeFormat"
			if col.Type == model.ColumnType_TIME ||
				col.Type == model.ColumnType_TIMETZ {
				tFormat = "HourMinuteSecondFormat"
			} else if col.Type == model.ColumnType_DATE {
				tFormat = "DateFormat"
			}

			bufPrepare.WriteString(fmt.Sprintf("var %s *time.Time\n",
				helper.GetVarName(col.Name)))
			bufPrepare.WriteString(fmt.Sprintf("if req.%s != nil {\n",
				helper.GetStructName(col.Name),
			))
			bufPrepare.WriteString(fmt.Sprintf("t, err := time.ParseInLocation(constraint.%s, *req.%s, time.Local)\n",
				tFormat,
				helper.GetStructName(col.Name),
			))
			bufPrepare.WriteString("if err != nil {\n")
			bufPrepare.WriteString(`log.WithError(err).Errorf("fail to parse time: %s", *req.` +
				helper.GetStructName(col.Name) +
				")\n")
			bufPrepare.WriteString("return nil, cerror.InvalidArgument(\"" + helper.GetStructName(col.Name) + "\")\n")
			bufPrepare.WriteString("}\n")
			bufPrepare.WriteString(fmt.Sprintf("%s = &t\n", helper.GetVarName(col.Name)))
			bufPrepare.WriteString("}\n")
			buf.WriteString(fmt.Sprintf("\n%s: %s,",
				helper.GetTableColName(col.Name),
				helper.GetVarName(col.Name)))
		} else {
			pt, _ := helper.GetProto3ValueType(col)
			pt, _ = helper.Proto3ValueType2GoType(pt)
			gt, _ := helper.GetGoType(col)
			if gt != pt {
				if !helper.IsGoTypeNullable(gt) {
					bufPrepare.WriteString(fmt.Sprintf("var %s *%s\n",
						helper.GetVarName(col.Name), gt))
					bufPrepare.WriteString(fmt.Sprintf("if req.%s != nil {\n",
						helper.GetStructName(col.Name),
					))
					bufPrepare.WriteString(fmt.Sprintf("*%s = %s(*req.%s)\n",
						helper.GetVarName(col.Name),
						gt,
						helper.GetStructName(col.Name),
					))
					bufPrepare.WriteString("}\n")
					buf.WriteString(fmt.Sprintf("\n%s: %s,",
						helper.GetTableColName(col.Name),
						helper.GetVarName(col.Name)))
				} else {
					buf.WriteString(fmt.Sprintf("\n%s: %s(req.%s),",
						helper.GetTableColName(col.Name),
						gt,
						helper.GetStructName(col.Name),
					))
				}
			} else {
				buf.WriteString(fmt.Sprintf("\n%s: req.%s,",
					helper.GetTableColName(col.Name),
					helper.GetStructName(col.Name),
				))
			}
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_UPDATE_TO_SET_GRPC, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_UPDATE_TO_SET_GRPC, buf.String())
}

// genRLDetailStructsAndFields 为主表生成RL表相关的结构体和字段
func genRLDetailStructsAndFields(code *string, mainTable *model.Table) {
	// 获取该主表的所有RL表
	rlTables := getRLTablesForMainTable(mainTable)

	// 1. 生成RL表Detail结构体定义
	var rlDetailBuf strings.Builder
	for _, rlTable := range rlTables {
		rlDetailBuf.WriteString(generateRLStructDefinition(rlTable, "Detail", false))
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_DETAIL_STRUCTS, rlDetailBuf.String())

	// 1.1 生成RL表Request结构体定义
	var rlRequestBuf strings.Builder
	for _, rlTable := range rlTables {
		rlRequestBuf.WriteString(generateRLStructDefinition(rlTable, "Request", false))
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_REQUEST_STRUCTS, rlRequestBuf.String())

	// 1.2 生成RL表ListInfo结构体定义（只包含list=true字段）
	var rlListInfoBuf strings.Builder
	for _, rlTable := range rlTables {
		// 检查是否有list=true的字段
		if !hasListFields(rlTable) {
			continue // 如果没有list字段，跳过
		}
		rlListInfoBuf.WriteString(generateRLStructDefinition(rlTable, "ListInfo", true))
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_LISTINFO_STRUCTS, rlListInfoBuf.String())

	// 2. 生成主表Detail中的RL字段
	var fieldsBuf strings.Builder
	for _, rlTable := range rlTables {
		fieldsBuf.WriteString(fmt.Sprintf("\n\t%ss []*%sDetail `json:\"%ss\"` // %s列表",
			helper.GetStructName(rlTable.Name),
			helper.GetStructName(rlTable.Name),
			helper.GetDirName(rlTable.Name),
			rlTable.Comment))
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_DETAIL, fieldsBuf.String())

	// 3. 生成ToDetail函数中的RL转换逻辑
	var convertBuf strings.Builder
	for _, rlTable := range rlTables {
		rlVarName := helper.GetVarName(rlTable.Name)
		rlStructName := helper.GetStructName(rlTable.Name)
		convertBuf.WriteString(fmt.Sprintf("var %ss []*%sDetail\n\t", rlVarName, rlStructName))
		convertBuf.WriteString(fmt.Sprintf("for _, %sBO := range bo.%ss {\n\t\t", rlVarName, rlStructName))
		convertBuf.WriteString(fmt.Sprintf("%sDetail := &%sDetail{", rlVarName, rlStructName))

		// 为每个RL表字段生成转换逻辑
		convertBuf.WriteString(generateRLFieldAssignments(rlTable, fmt.Sprintf("%sBO", rlVarName), false, false))

		convertBuf.WriteString("\n\t}\n\t\t")
		convertBuf.WriteString(fmt.Sprintf("%ss = append(%ss, %sDetail)\n\t", rlVarName, rlVarName, rlVarName))
		convertBuf.WriteString("}\n\t")
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_CONVERT_IN_TO_DETAIL, convertBuf.String())

	// 4. 生成主表Detail赋值中的RL字段
	var assignBuf strings.Builder
	for _, rlTable := range rlTables {
		rlVarName := helper.GetVarName(rlTable.Name)
		rlStructName := helper.GetStructName(rlTable.Name)
		assignBuf.WriteString(fmt.Sprintf("\n\t\t%ss: %ss,", rlStructName, rlVarName))
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_ASSIGN_IN_DETAIL, assignBuf.String())

	// 5. 生成主表ReqCreate中的RL字段
	var createFieldsBuf strings.Builder
	for _, rlTable := range rlTables {
		rlStructName := helper.GetStructName(rlTable.Name)
		createFieldsBuf.WriteString(fmt.Sprintf("\n\t%ss []ReqCreate%s `json:\"%ss\"` // %s列表",
			rlStructName,
			rlStructName,
			helper.GetDirName(rlTable.Name),
			rlTable.Comment))
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_CREATE, createFieldsBuf.String())

	// 6. 生成Create转BO时的RL字段赋值
	var createAssignBuf strings.Builder
	for _, rlTable := range rlTables {
		rlVarName := helper.GetVarName(rlTable.Name)
		rlStructName := helper.GetStructName(rlTable.Name)
		createAssignBuf.WriteString(fmt.Sprintf("var %ss []*biz.%sBO\n\t", rlVarName, rlStructName))
		createAssignBuf.WriteString(fmt.Sprintf("for _, %sData := range req.%ss {\n\t\t", rlVarName, rlStructName))
		createAssignBuf.WriteString(fmt.Sprintf("%sBO := &biz.%sBO{\n", rlVarName, rlStructName))

		// 为每个RL表的alter=true字段生成赋值
		for _, col := range rlTable.Columns {
			if !col.IsAlterable {
				continue
			}
			if col.IsHidden {
				continue
			}
			createAssignBuf.WriteString(fmt.Sprintf("\t\t\t%s: %sData.%s,\n",
				helper.GetTableColName(col.Name),
				rlVarName,
				helper.GetStructName(col.Name)))
		}

		createAssignBuf.WriteString("\t\t}\n\t\t")
		createAssignBuf.WriteString(fmt.Sprintf("%ss = append(%ss, %sBO)\n\t", rlVarName, rlVarName, rlVarName))
		createAssignBuf.WriteString("}\n\t")
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_CREATE_ASSIGN_TO_BO, createAssignBuf.String())

	// 7. 在主表BO的赋值中添加RL字段 - 这部分由genAssignCreateToBO函数处理
	// 不需要在这里处理，因为会与genAssignCreateToBO冲突

	// 8. 生成主表ListInfo中的RL字段（只对有list字段的RL表）
	var rlListInfoFieldsBuf strings.Builder
	for _, rlTable := range rlTables {
		// 检查是否有list=true的字段
		if !hasListFields(rlTable) {
			continue // 如果没有list字段，跳过
		}

		rlStructName := helper.GetStructName(rlTable.Name)
		rlListInfoFieldsBuf.WriteString(fmt.Sprintf("\n\t%ss []*%sListInfo `json:\"%ss\"` // %s列表",
			rlStructName,
			rlStructName,
			helper.GetDirName(rlTable.Name),
			rlTable.Comment))
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_IN_LISTINFO, rlListInfoFieldsBuf.String())

	// 9. 生成ToListInfo函数中的RL转换逻辑
	var rlListInfoConvertBuf strings.Builder
	for _, rlTable := range rlTables {
		// 检查是否有list=true的字段
		if !hasListFields(rlTable) {
			continue // 如果没有list字段，跳过
		}

		rlVarName := helper.GetVarName(rlTable.Name)
		rlStructName := helper.GetStructName(rlTable.Name)

		rlListInfoConvertBuf.WriteString(fmt.Sprintf("var %ss []*%sListInfo\n\t\t", rlVarName, rlStructName))
		rlListInfoConvertBuf.WriteString(fmt.Sprintf("for _, %sBO := range objs[i].%ss {\n\t\t\t", rlVarName, rlStructName))

		// 生成字段处理逻辑
		prepareCode, assignCode := generateRLListInfoFieldLogic(rlTable, rlVarName)

		if prepareCode != "" {
			rlListInfoConvertBuf.WriteString(prepareCode)
		}

		rlListInfoConvertBuf.WriteString(fmt.Sprintf("%sListInfo := &%sListInfo{%s\n\t\t\t",
			rlVarName, rlStructName, assignCode))
		rlListInfoConvertBuf.WriteString("}\n\t\t\t")
		rlListInfoConvertBuf.WriteString(fmt.Sprintf("%ss = append(%ss, %sListInfo)\n\t\t", rlVarName, rlVarName, rlVarName))
		rlListInfoConvertBuf.WriteString("}\n\t\t")
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_CONVERT_IN_TO_LISTINFO, rlListInfoConvertBuf.String())

	// 10. 生成ToListInfo函数中的RL字段赋值
	var rlListInfoAssignBuf strings.Builder
	for _, rlTable := range rlTables {
		// 检查是否有list=true的字段
		if !hasListFields(rlTable) {
			continue // 如果没有list字段，跳过
		}

		rlVarName := helper.GetVarName(rlTable.Name)
		rlStructName := helper.GetStructName(rlTable.Name)
		rlListInfoAssignBuf.WriteString(fmt.Sprintf("\n\t\t\t%ss: %ss,", rlStructName, rlVarName))
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_ASSIGN_IN_LISTINFO, rlListInfoAssignBuf.String())

	// 11. 生成RL表操作函数（Add、Remove、Get）
	var rlHandlerFuncsBuf strings.Builder
	for _, rlTable := range rlTables {
		mainTableStructName := helper.GetStructName(mainTable.Name)
		mainTableVarName := helper.GetVarName(mainTable.Name)
		rlTableStructName := helper.GetStructName(rlTable.Name)
		rlTableVarName := helper.GetVarName(rlTable.Name)

		// Add函数
		addFunc := template.TplRLHandlerAdd
		addFunc = strings.ReplaceAll(addFunc, template.PH_TABLE_NAME_STRUCT, mainTableStructName)
		addFunc = strings.ReplaceAll(addFunc, template.PH_TABLE_NAME_VAR, mainTableVarName)
		addFunc = strings.ReplaceAll(addFunc, template.PH_TABLE_NAME_URI, helper.GetURIName(mainTable.Name))
		addFunc = strings.ReplaceAll(addFunc, template.PH_TABLE_COMMENT, mainTable.Comment)
		addFunc = strings.ReplaceAll(addFunc, template.PH_RL_TABLE_NAME_STRUCT, rlTableStructName)
		addFunc = strings.ReplaceAll(addFunc, template.PH_RL_TABLE_NAME_VAR, rlTableVarName)
		addFunc = strings.ReplaceAll(addFunc, template.PH_RL_TABLE_NAME_URI, helper.GetURIName(rlTable.Name))
		addFunc = strings.ReplaceAll(addFunc, template.PH_RL_TABLE_COMMENT, rlTable.Comment)

		// 生成BO字段赋值
		boAssignments := generateRLHandlerFieldAssignments(rlTable, rlTableVarName, "bo_assign")
		addFunc = strings.ReplaceAll(addFunc, template.PH_RL_BO_ASSIGN, boAssignments)

		// 生成Detail字段赋值
		detailAssignments := generateRLHandlerFieldAssignments(rlTable, rlTableVarName, "detail_assign")
		addFunc = strings.ReplaceAll(addFunc, template.PH_RL_DETAIL_ASSIGN, detailAssignments)
		rlHandlerFuncsBuf.WriteString(addFunc)

		// Remove函数
		removeFunc := template.TplRLHandlerRemove
		removeFunc = strings.ReplaceAll(removeFunc, template.PH_TABLE_NAME_STRUCT, mainTableStructName)
		removeFunc = strings.ReplaceAll(removeFunc, template.PH_TABLE_NAME_VAR, mainTableVarName)
		removeFunc = strings.ReplaceAll(removeFunc, template.PH_TABLE_NAME_URI, helper.GetURIName(mainTable.Name))
		removeFunc = strings.ReplaceAll(removeFunc, template.PH_TABLE_COMMENT, mainTable.Comment)
		removeFunc = strings.ReplaceAll(removeFunc, template.PH_RL_TABLE_NAME_STRUCT, rlTableStructName)
		removeFunc = strings.ReplaceAll(removeFunc, template.PH_RL_TABLE_NAME_VAR, rlTableVarName)
		removeFunc = strings.ReplaceAll(removeFunc, template.PH_RL_TABLE_NAME_URI, helper.GetURIName(rlTable.Name))
		removeFunc = strings.ReplaceAll(removeFunc, template.PH_RL_TABLE_COMMENT, rlTable.Comment)
		rlHandlerFuncsBuf.WriteString(removeFunc)

		// Get函数
		getFunc := template.TplRLHandlerGet
		getFunc = strings.ReplaceAll(getFunc, template.PH_TABLE_NAME_STRUCT, mainTableStructName)
		getFunc = strings.ReplaceAll(getFunc, template.PH_TABLE_NAME_VAR, mainTableVarName)
		getFunc = strings.ReplaceAll(getFunc, template.PH_TABLE_NAME_URI, helper.GetURIName(mainTable.Name))
		getFunc = strings.ReplaceAll(getFunc, template.PH_TABLE_COMMENT, mainTable.Comment)
		getFunc = strings.ReplaceAll(getFunc, template.PH_RL_TABLE_NAME_STRUCT, rlTableStructName)
		getFunc = strings.ReplaceAll(getFunc, template.PH_RL_TABLE_NAME_VAR, rlTableVarName)
		getFunc = strings.ReplaceAll(getFunc, template.PH_RL_TABLE_NAME_URI, helper.GetURIName(rlTable.Name))
		getFunc = strings.ReplaceAll(getFunc, template.PH_RL_TABLE_COMMENT, rlTable.Comment)

		// 生成Detail字段赋值（循环中）
		detailAssignLoopAssignments := generateRLHandlerFieldAssignments(rlTable, rlTableVarName, "detail_assign_loop")
		getFunc = strings.ReplaceAll(getFunc, template.PH_RL_DETAIL_ASSIGN_LOOP, detailAssignLoopAssignments)
		rlHandlerFuncsBuf.WriteString(getFunc)
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_HANDLER_FUNCTIONS, rlHandlerFuncsBuf.String())
}

// genRLGRPCHandlerFunctions 为gRPC生成RL表操作的相关代码
func genRLGRPCHandlerFunctions(code *string, mainTable *model.Table) {
	// 获取该主表的所有RL表
	rlTables := getRLTablesForMainTable(mainTable)

	// 1. 生成gRPC ToDetail函数中的RL转换逻辑
	var convertBuf strings.Builder
	for _, rlTable := range rlTables {
		rlVarName := helper.GetVarName(rlTable.Name)
		rlStructName := helper.GetStructName(rlTable.Name)
		convertBuf.WriteString(fmt.Sprintf("var %ss []*api.%sDetail\n\t", rlVarName, rlStructName))
		convertBuf.WriteString(fmt.Sprintf("for _, %sBO := range bo.%ss {\n\t\t", rlVarName, rlStructName))
		convertBuf.WriteString(fmt.Sprintf("%sDetail := &api.%sDetail{", rlVarName, rlStructName))

		// 为每个RL表字段生成转换逻辑
		convertBuf.WriteString(generateGRPCRLFieldAssignments(rlTable, fmt.Sprintf("%sBO", rlVarName), false))

		convertBuf.WriteString("\n\t}\n\t\t")
		convertBuf.WriteString(fmt.Sprintf("%ss = append(%ss, %sDetail)\n\t", rlVarName, rlVarName, rlVarName))
		convertBuf.WriteString("}\n\t")
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_CONVERT_IN_TO_DETAIL_GRPC, convertBuf.String())

	// 2. 生成gRPC主表Detail赋值中的RL字段
	var assignBuf strings.Builder
	for _, rlTable := range rlTables {
		rlVarName := helper.GetVarName(rlTable.Name)
		rlStructName := helper.GetStructName(rlTable.Name)
		assignBuf.WriteString(fmt.Sprintf("\n\t\t%ss: %ss,", rlStructName, rlVarName))
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_ASSIGN_IN_DETAIL_GRPC, assignBuf.String())

	// 3. 生成gRPC Create转BO时的RL字段赋值
	var createAssignBuf strings.Builder
	for _, rlTable := range rlTables {
		rlVarName := helper.GetVarName(rlTable.Name)
		rlStructName := helper.GetStructName(rlTable.Name)
		createAssignBuf.WriteString(fmt.Sprintf("var %ss []*biz.%sBO\n\t", rlVarName, rlStructName))
		createAssignBuf.WriteString(fmt.Sprintf("for _, %sData := range req.%ss {\n\t\t", rlVarName, rlStructName))
		createAssignBuf.WriteString(fmt.Sprintf("%sBO := &biz.%sBO{", rlVarName, rlStructName))

		// 为每个RL表的alter=true字段生成赋值
		createAssignBuf.WriteString(generateGRPCCreateBOAssignments(rlTable, rlVarName))

		createAssignBuf.WriteString("\n\t}\n\t\t")
		createAssignBuf.WriteString(fmt.Sprintf("%ss = append(%ss, %sBO)\n\t", rlVarName, rlVarName, rlVarName))
		createAssignBuf.WriteString("}\n\t")
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_CREATE_ASSIGN_TO_BO_GRPC, createAssignBuf.String())

	// 4. 生成gRPC ToListInfo函数中的RL转换逻辑
	var rlListInfoConvertBuf strings.Builder
	for _, rlTable := range rlTables {
		// 检查是否有list=true的字段
		if !hasListFields(rlTable) {
			continue // 如果没有list字段，跳过
		}

		rlVarName := helper.GetVarName(rlTable.Name)
		rlStructName := helper.GetStructName(rlTable.Name)

		rlListInfoConvertBuf.WriteString(fmt.Sprintf("var %ss []*api.%sListInfo\n\t\t", rlVarName, rlStructName))
		rlListInfoConvertBuf.WriteString(fmt.Sprintf("for _, %sBO := range objs[i].%ss {\n\t\t\t", rlVarName, rlStructName))

		// 生成字段赋值逻辑
		assignFields := generateGRPCRLFieldAssignments(rlTable, fmt.Sprintf("%sBO", rlVarName), true)

		rlListInfoConvertBuf.WriteString(fmt.Sprintf("%sListInfo := &api.%sListInfo{%s\n\t\t\t",
			rlVarName, rlStructName, assignFields))
		rlListInfoConvertBuf.WriteString("}\n\t\t\t")
		rlListInfoConvertBuf.WriteString(fmt.Sprintf("%ss = append(%ss, %sListInfo)\n\t\t", rlVarName, rlVarName, rlVarName))
		rlListInfoConvertBuf.WriteString("}\n\t\t")
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_CONVERT_IN_TO_LISTINFO_GRPC, rlListInfoConvertBuf.String())

	// 5. 生成gRPC ToListInfo函数中的RL字段赋值
	var rlListInfoAssignBuf strings.Builder
	for _, rlTable := range rlTables {
		// 检查是否有list=true的字段
		if !hasListFields(rlTable) {
			continue // 如果没有list字段，跳过
		}

		rlVarName := helper.GetVarName(rlTable.Name)
		rlStructName := helper.GetStructName(rlTable.Name)
		rlListInfoAssignBuf.WriteString(fmt.Sprintf("\n\t\t\t%ss: %ss,", rlStructName, rlVarName))
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_FIELDS_ASSIGN_IN_LISTINFO_GRPC, rlListInfoAssignBuf.String())

	// 6. 生成gRPC RL表操作函数（Add、Remove、Get）
	var rlGrpcHandlerFuncsBuf strings.Builder
	for _, rlTable := range rlTables {
		mainTableStructName := helper.GetStructName(mainTable.Name)
		rlTableStructName := helper.GetStructName(rlTable.Name)
		rlTableVarName := helper.GetVarName(rlTable.Name)

		// Add函数
		addFunc := template.TplGRPCRLHandlerAdd
		addFunc = strings.ReplaceAll(addFunc, template.PH_TABLE_NAME_STRUCT, mainTableStructName)
		addFunc = strings.ReplaceAll(addFunc, template.PH_RL_TABLE_NAME_STRUCT, rlTableStructName)
		addFunc = strings.ReplaceAll(addFunc, template.PH_RL_TABLE_NAME_VAR, rlTableVarName)
		addFunc = strings.ReplaceAll(addFunc, template.PH_RL_TABLE_COMMENT, rlTable.Comment)

		// 生成gRPC BO字段赋值
		grpcBoAssignments := generateGRPCRLHandlerFieldAssignments(rlTable, rlTableVarName, "grpc_bo_assign")
		addFunc = strings.ReplaceAll(addFunc, template.PH_RL_BO_ASSIGN_GRPC, grpcBoAssignments)

		// 生成gRPC Detail字段赋值
		grpcDetailAssignments := generateGRPCRLHandlerFieldAssignments(rlTable, rlTableVarName, "grpc_detail_assign")
		addFunc = strings.ReplaceAll(addFunc, template.PH_RL_DETAIL_ASSIGN_GRPC, grpcDetailAssignments)
		rlGrpcHandlerFuncsBuf.WriteString(addFunc)

		// Remove函数
		removeFunc := template.TplGRPCRLHandlerRemove
		removeFunc = strings.ReplaceAll(removeFunc, template.PH_TABLE_NAME_STRUCT, mainTableStructName)
		removeFunc = strings.ReplaceAll(removeFunc, template.PH_RL_TABLE_NAME_STRUCT, rlTableStructName)
		removeFunc = strings.ReplaceAll(removeFunc, template.PH_RL_TABLE_COMMENT, rlTable.Comment)
		rlGrpcHandlerFuncsBuf.WriteString(removeFunc)

		// Get函数
		getFunc := template.TplGRPCRLHandlerGet
		getFunc = strings.ReplaceAll(getFunc, template.PH_TABLE_NAME_STRUCT, mainTableStructName)
		getFunc = strings.ReplaceAll(getFunc, template.PH_RL_TABLE_NAME_STRUCT, rlTableStructName)
		getFunc = strings.ReplaceAll(getFunc, template.PH_RL_TABLE_NAME_VAR, rlTableVarName)
		getFunc = strings.ReplaceAll(getFunc, template.PH_RL_TABLE_COMMENT, rlTable.Comment)

		// 生成gRPC Detail字段赋值（循环中）
		grpcDetailAssignLoopAssignments := generateGRPCRLHandlerFieldAssignments(rlTable, rlTableVarName, "grpc_detail_assign_loop")
		getFunc = strings.ReplaceAll(getFunc, template.PH_RL_DETAIL_ASSIGN_LOOP_GRPC, grpcDetailAssignLoopAssignments)
		rlGrpcHandlerFuncsBuf.WriteString(getFunc)
	}
	*code = strings.ReplaceAll(*code, template.PH_RL_GRPC_HANDLER_FUNCTIONS, rlGrpcHandlerFuncsBuf.String())
}

// getRLTablesForMainTable 获取指定主表的所有RL表
func getRLTablesForMainTable(mainTable *model.Table) []*model.Table {
	project := mainTable.Database.Project

	// 构建表名到表的映射
	tableNameToTable := make(map[string]*model.Table)
	for _, table := range project.Database.Tables {
		tableNameToTable[table.Name] = table
	}

	// 获取该主表的所有RL表
	var rlTables []*model.Table
	for _, table := range project.Database.Tables {
		if table.Type == model.TableType_RL {
			// 查找指向主表的外键
			identifiedMainTable := helper.IdentifyRLMainTable(table, tableNameToTable)
			if identifiedMainTable != nil && identifiedMainTable.Name == mainTable.Name {
				rlTables = append(rlTables, table)
			}
		}
	}
	return rlTables
}

// getTimeFormat 根据列类型获取时间格式
func getTimeFormat(colType model.ColumnType) string {
	switch colType {
	case model.ColumnType_TIME, model.ColumnType_TIMETZ:
		return "HourMinuteSecondFormat"
	case model.ColumnType_DATE:
		return "DateFormat"
	default:
		return "SecondTimeFormat"
	}
}

// isTimeColumn 判断是否为时间类型列
func isTimeColumn(col *model.Column) bool {
	return col.Type == model.ColumnType_DATETIME ||
		col.Type == model.ColumnType_TIMESTAMP ||
		col.Type == model.ColumnType_TIME ||
		col.Type == model.ColumnType_DATE ||
		col.Type == model.ColumnType_TIMETZ ||
		col.Type == model.ColumnType_TIMESTAMPTZ
}

// hasListFields 检查RL表是否有List字段
func hasListFields(rlTable *model.Table) bool {
	for _, col := range rlTable.Columns {
		if !col.IsHidden && col.IsList {
			return true
		}
	}
	return false
}

// generateTimeFieldAssignment 生成时间字段的赋值代码
func generateTimeFieldAssignment(col *model.Column, sourcePrefix, targetField string, isGRPC bool) string {
	if !isTimeColumn(col) {
		return ""
	}

	tFormat := getTimeFormat(col.Type)
	sourceField := fmt.Sprintf("%s.%s", sourcePrefix, helper.GetTableColName(col.Name))

	if isGRPC {
		if !col.IsRequired {
			return fmt.Sprintf("\n\t\t\t%s: func() *string { if %s != nil { t := jgstr.FormatTime(*%s); return &t }; return nil }(),",
				targetField, sourceField, sourceField)
		} else {
			return fmt.Sprintf("\n\t\t\t%s: jgstr.FormatTime(%s),", targetField, sourceField)
		}
	} else {
		if !col.IsRequired {
			// 对于nullable字段，先返回nil，后续需要特殊处理
			return fmt.Sprintf("\n\t\t\t%s: nil, // TODO: 处理nullable时间字段的转换", targetField)
		} else {
			return fmt.Sprintf("\n\t\t\t%s: %s.Format(constraint.%s),", targetField, sourceField, tFormat)
		}
	}
}

// generateRLFieldAssignments 生成RL字段转换的完整逻辑
func generateRLFieldAssignments(rlTable *model.Table, sourcePrefix string, onlyListFields bool, isGRPC bool) string {
	var buf strings.Builder

	for _, col := range rlTable.Columns {
		if col.IsHidden {
			continue
		}
		if onlyListFields && !col.IsList {
			continue
		}

		targetField := helper.GetStructName(col.Name)
		if isTimeColumn(col) {
			buf.WriteString(generateTimeFieldAssignment(col, sourcePrefix, targetField, isGRPC))
		} else {
			buf.WriteString(fmt.Sprintf("\n\t\t\t%s: %s.%s,", targetField, sourcePrefix, helper.GetTableColName(col.Name)))
		}
	}

	return buf.String()
}

// generateRLListInfoFieldLogic 生成ToListInfo的复杂字段处理逻辑
func generateRLListInfoFieldLogic(rlTable *model.Table, rlVarName string) (prepareCode, assignCode string) {
	var prepareTimeFields, assignFields strings.Builder

	for _, col := range rlTable.Columns {
		if col.IsHidden || !col.IsList {
			continue
		}

		targetField := helper.GetStructName(col.Name)
		sourceField := fmt.Sprintf("%sBO.%s", rlVarName, helper.GetTableColName(col.Name))

		if isTimeColumn(col) {
			tFormat := getTimeFormat(col.Type)

			if !col.IsRequired {
				varName := helper.GetVarName(col.Name)
				prepareTimeFields.WriteString(fmt.Sprintf("var %s *string\n\t\t\t", varName))
				prepareTimeFields.WriteString(fmt.Sprintf("if %s != nil {\n\t\t\t\t", sourceField))
				prepareTimeFields.WriteString(fmt.Sprintf("*%s = %s.Format(constraint.%s)\n\t\t\t", varName, sourceField, tFormat))
				prepareTimeFields.WriteString("}\n\t\t\t")
				assignFields.WriteString(fmt.Sprintf("\n\t\t\t\t%s: %s,", targetField, varName))
			} else {
				assignFields.WriteString(fmt.Sprintf("\n\t\t\t\t%s: %s.Format(constraint.%s),", targetField, sourceField, tFormat))
			}
		} else {
			assignFields.WriteString(fmt.Sprintf("\n\t\t\t\t%s: %s,", targetField, sourceField))
		}
	}

	return prepareTimeFields.String(), assignFields.String()
}

// generateRLHandlerFieldAssignments 生成Handler函数中的字段赋值
func generateRLHandlerFieldAssignments(rlTable *model.Table, rlTableVarName string, assignmentType string) string {
	var buf strings.Builder

	for _, col := range rlTable.Columns {
		if col.IsHidden {
			continue
		}

		switch assignmentType {
		case "bo_assign":
			if !col.IsAlterable {
				continue
			}
			buf.WriteString(fmt.Sprintf("\n\t\t%s: req.%s,",
				helper.GetTableColName(col.Name),
				helper.GetStructName(col.Name)))
		case "detail_assign":
			targetField := helper.GetStructName(col.Name)
			sourceField := fmt.Sprintf("%sBO.%s", rlTableVarName, helper.GetTableColName(col.Name))

			if isTimeColumn(col) {
				tFormat := getTimeFormat(col.Type)
				if !col.IsRequired {
					buf.WriteString(fmt.Sprintf("\n\t\t%s: func() *string { if %s != nil { t := %s.Format(constraint.%s); return &t }; return nil }(),",
						targetField, sourceField, sourceField, tFormat))
				} else {
					buf.WriteString(fmt.Sprintf("\n\t\t%s: %s.Format(constraint.%s),",
						targetField, sourceField, tFormat))
				}
			} else {
				buf.WriteString(fmt.Sprintf("\n\t\t%s: %s,", targetField, sourceField))
			}
		case "detail_assign_loop":
			targetField := helper.GetStructName(col.Name)
			sourceField := fmt.Sprintf("%sBO.%s", rlTableVarName, helper.GetTableColName(col.Name))

			if isTimeColumn(col) {
				tFormat := getTimeFormat(col.Type)
				if !col.IsRequired {
					buf.WriteString(fmt.Sprintf("\n\t\t\t%s: func() *string { if %s != nil { t := %s.Format(constraint.%s); return &t }; return nil }(),",
						targetField, sourceField, sourceField, tFormat))
				} else {
					buf.WriteString(fmt.Sprintf("\n\t\t\t%s: %s.Format(constraint.%s),",
						targetField, sourceField, tFormat))
				}
			} else {
				buf.WriteString(fmt.Sprintf("\n\t\t\t%s: %s,", targetField, sourceField))
			}
		}
	}

	return buf.String()
}

// generateGRPCRLFieldAssignments 生成gRPC RL字段转换的完整逻辑
func generateGRPCRLFieldAssignments(rlTable *model.Table, sourcePrefix string, onlyListFields bool) string {
	var buf strings.Builder

	for _, col := range rlTable.Columns {
		if col.IsHidden {
			continue
		}
		if onlyListFields && !col.IsList {
			continue
		}

		targetField := helper.GetStructName(col.Name)
		sourceField := fmt.Sprintf("%s.%s", sourcePrefix, helper.GetTableColName(col.Name))

		if isTimeColumn(col) {
			if !col.IsRequired {
				buf.WriteString(fmt.Sprintf("\n\t\t\t%s: func() *string { if %s != nil { t := jgstr.FormatTime(*%s); return &t }; return nil }(),",
					targetField, sourceField, sourceField))
			} else {
				buf.WriteString(fmt.Sprintf("\n\t\t\t%s: jgstr.FormatTime(%s),", targetField, sourceField))
			}
		} else {
			buf.WriteString(fmt.Sprintf("\n\t\t\t%s: %s,", targetField, sourceField))
		}
	}

	return buf.String()
}

// generateGRPCCreateBOAssignments 生成gRPC Create转BO的字段赋值
func generateGRPCCreateBOAssignments(rlTable *model.Table, rlVarName string) string {
	var buf strings.Builder

	for _, col := range rlTable.Columns {
		if !col.IsAlterable || col.IsHidden {
			continue
		}

		if isTimeColumn(col) {
			if !col.IsRequired {
				buf.WriteString(fmt.Sprintf("\n\t\t\t%s: func() *time.Time { if %sData.%s != nil { t := jgstr.ParseTime(*%sData.%s); return &t }; return nil }(),",
					helper.GetTableColName(col.Name),
					rlVarName,
					helper.GetStructName(col.Name),
					rlVarName,
					helper.GetStructName(col.Name)))
			} else {
				buf.WriteString(fmt.Sprintf("\n\t\t\t%s: jgstr.ParseTime(%sData.%s),",
					helper.GetTableColName(col.Name),
					rlVarName,
					helper.GetStructName(col.Name)))
			}
		} else {
			buf.WriteString(fmt.Sprintf("\n\t\t\t%s: %sData.%s,",
				helper.GetTableColName(col.Name),
				rlVarName,
				helper.GetStructName(col.Name)))
		}
	}

	return buf.String()
}

// generateGRPCRLHandlerFieldAssignments 生成gRPC Handler函数中的字段赋值
func generateGRPCRLHandlerFieldAssignments(rlTable *model.Table, rlTableVarName string, assignmentType string) string {
	var buf strings.Builder

	for _, col := range rlTable.Columns {
		if col.IsHidden {
			continue
		}

		switch assignmentType {
		case "grpc_bo_assign":
			if !col.IsAlterable {
				continue
			}

			if isTimeColumn(col) {
				if !col.IsRequired {
					buf.WriteString(fmt.Sprintf("\n\t\t%s: func() *time.Time { if req.%s != nil { t := jgstr.ParseTime(*req.%s); return &t }; return nil }(),",
						helper.GetTableColName(col.Name),
						helper.GetStructName(col.Name),
						helper.GetStructName(col.Name)))
				} else {
					buf.WriteString(fmt.Sprintf("\n\t\t%s: jgstr.ParseTime(req.%s),",
						helper.GetTableColName(col.Name),
						helper.GetStructName(col.Name)))
				}
			} else {
				buf.WriteString(fmt.Sprintf("\n\t\t%s: req.%s,",
					helper.GetTableColName(col.Name),
					helper.GetStructName(col.Name)))
			}
		case "grpc_detail_assign":
			targetField := helper.GetStructName(col.Name)
			sourceField := fmt.Sprintf("%sBO.%s", rlTableVarName, helper.GetTableColName(col.Name))

			if isTimeColumn(col) {
				if !col.IsRequired {
					buf.WriteString(fmt.Sprintf("\n\t\t%s: func() *string { if %s != nil { t := jgstr.FormatTime(*%s); return &t }; return nil }(),",
						targetField, sourceField, sourceField))
				} else {
					buf.WriteString(fmt.Sprintf("\n\t\t%s: jgstr.FormatTime(%s),",
						targetField, sourceField))
				}
			} else {
				buf.WriteString(fmt.Sprintf("\n\t\t%s: %s,", targetField, sourceField))
			}
		case "grpc_detail_assign_loop":
			targetField := helper.GetStructName(col.Name)
			sourceField := fmt.Sprintf("%sBO.%s", rlTableVarName, helper.GetTableColName(col.Name))

			if isTimeColumn(col) {
				if !col.IsRequired {
					buf.WriteString(fmt.Sprintf("\n\t\t\t%s: func() *string { if %s != nil { t := jgstr.FormatTime(*%s); return &t }; return nil }(),",
						targetField, sourceField, sourceField))
				} else {
					buf.WriteString(fmt.Sprintf("\n\t\t\t%s: jgstr.FormatTime(%s),",
						targetField, sourceField))
				}
			} else {
				buf.WriteString(fmt.Sprintf("\n\t\t\t%s: %s,", targetField, sourceField))
			}
		}
	}

	return buf.String()
}

// generateRLStructFields 生成RL结构体的字段定义
func generateRLStructFields(rlTable *model.Table, onlyListFields bool) string {
	var buf strings.Builder

	for _, col := range rlTable.Columns {
		if col.IsHidden {
			continue
		}
		if onlyListFields && !col.IsList {
			continue
		}

		goType, err := helper.GetGoType(col)
		if err != nil {
			log.Fatalf("fail to get go type: %v", err)
		}

		// 处理时间类型
		if isTimeColumn(col) {
			goType = "string"
		}

		if !col.IsRequired && !helper.IsGoTypeNullable(goType) {
			goType = "*" + goType
		}

		comment := col.Comment
		if !col.IsRequired {
			comment += " (nullable)"
		}

		buf.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"` // %s\n",
			helper.GetStructName(col.Name),
			goType,
			helper.GetDirName(col.Name),
			comment))
	}

	return buf.String()
}

// generateRLFieldConversion 生成RL字段转换逻辑
func generateRLFieldConversion(col *model.Column, sourcePrefix, targetPrefix string, isGRPC bool) string {
	targetField := helper.GetStructName(col.Name)

	if isTimeColumn(col) {
		return generateTimeFieldAssignment(col, sourcePrefix, targetField, isGRPC)
	} else {
		return fmt.Sprintf("\n\t\t\t%s: %s.%s,", targetField, sourcePrefix, helper.GetTableColName(col.Name))
	}
}

// generateRLStructDefinition 生成单个RL结构体定义
func generateRLStructDefinition(rlTable *model.Table, structType string, onlyListFields bool) string {
	var buf strings.Builder

	switch structType {
	case "Detail":
		buf.WriteString(fmt.Sprintf("// %s详情\n", rlTable.Comment))
		buf.WriteString(fmt.Sprintf("type %sDetail struct {\n", helper.GetStructName(rlTable.Name)))
	case "Request":
		buf.WriteString(fmt.Sprintf("// %s创建数据\n", rlTable.Comment))
		buf.WriteString(fmt.Sprintf("type ReqCreate%s struct {\n", helper.GetStructName(rlTable.Name)))
	case "ListInfo":
		buf.WriteString(fmt.Sprintf("// %s列表信息\n", rlTable.Comment))
		buf.WriteString(fmt.Sprintf("type %sListInfo struct {\n", helper.GetStructName(rlTable.Name)))
	}

	if structType == "Request" {
		// Request结构体使用不同的字段生成逻辑
		for _, col := range rlTable.Columns {
			if !col.IsAlterable || col.IsHidden {
				continue
			}
			goType := helper.GetGoTypeForHandler(col)
			comment := helper.GetCommentForHandler(col)
			binding := helper.GetBinding(col)
			buf.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"%s` // %s\n",
				helper.GetStructName(col.Name),
				goType,
				col.Name,
				binding,
				comment))
		}
	} else {
		buf.WriteString(generateRLStructFields(rlTable, onlyListFields))
	}

	buf.WriteString("}\n\n")
	return buf.String()
}

// genBRHTTPHandlerFunctions 为DATA表生成BR关系的HTTP handler函数
func genBRHTTPHandlerFunctions(code *string, mainTable *model.Table) {
	// 构建表名到表的映射
	tableNameToTable := make(map[string]*model.Table)
	for _, t := range mainTable.Database.Tables {
		tableNameToTable[t.Name] = t
	}

	// 获取当前表的所有BR表关系
	brTables := helper.GetMainTableBRs(mainTable, mainTable.Database.Tables)

	if len(brTables) == 0 {
		*code = strings.ReplaceAll(*code, template.PH_BR_HTTP_HANDLER_FUNCTIONS, "")
		return
	}

	var brHandlerFuncsBuf strings.Builder

	// 为每个BR表生成GET handler函数
	for _, brTable := range brTables {
		// 获取对方表
		otherTable := helper.GetBROtherTable(brTable, mainTable, tableNameToTable)
		if otherTable == nil {
			continue
		}

		// 生成handler函数
		getFunc := template.TplBRHandlerGet

		// 替换当前表相关的占位符
		getFunc = strings.ReplaceAll(getFunc, template.PH_TABLE_NAME_STRUCT, helper.GetStructName(mainTable.Name))
		getFunc = strings.ReplaceAll(getFunc, template.PH_TABLE_NAME_VAR, helper.GetVarName(mainTable.Name))
		getFunc = strings.ReplaceAll(getFunc, template.PH_TABLE_NAME_URI, helper.GetURIName(mainTable.Name))
		getFunc = strings.ReplaceAll(getFunc, template.PH_TABLE_COMMENT, mainTable.Comment)

		// 替换对方表相关的占位符
		getFunc = strings.ReplaceAll(getFunc, template.PH_OTHER_TABLE_NAME_STRUCT, helper.GetStructName(otherTable.Name))
		getFunc = strings.ReplaceAll(getFunc, template.PH_OTHER_TABLE_NAME_URI, helper.GetURIName(otherTable.Name))
		getFunc = strings.ReplaceAll(getFunc, template.PH_OTHER_TABLE_NAME_VAR, helper.GetVarName(otherTable.Name))
		getFunc = strings.ReplaceAll(getFunc, template.PH_OTHER_TABLE_COMMENT, otherTable.Comment)

		// 生成对方表的筛选字段文档
		otherFilterDoc := generateBRFilterDoc(otherTable)
		getFunc = strings.ReplaceAll(getFunc, template.PH_OTHER_COL_LIST_FOR_FILTER_DOC, otherFilterDoc)

		// 生成对方表的排序字段文档
		otherOrderDoc := generateBROrderDoc(otherTable)
		getFunc = strings.ReplaceAll(getFunc, template.PH_OTHER_COL_LIST_FOR_ORDER_DOC, otherOrderDoc)

		// 生成筛选条件准备赋值
		otherFilterPrepare := generateBRFilterPrepare(otherTable)
		getFunc = strings.ReplaceAll(getFunc, template.PH_OTHER_PREPARE_ASSIGN_FILTER_TO_OPTION, otherFilterPrepare)

		// 生成筛选条件赋值
		otherFilterAssign := generateBRFilterAssign(otherTable)
		getFunc = strings.ReplaceAll(getFunc, template.PH_OTHER_FILTER_ASSIGN_TO_OPTION, otherFilterAssign)

		// 生成排序条件赋值
		otherOrderAssign := generateBROrderAssign(otherTable)
		getFunc = strings.ReplaceAll(getFunc, template.PH_OTHER_ASSIGN_ORDER_TO_OPTION, otherOrderAssign)

		brHandlerFuncsBuf.WriteString(getFunc)
	}

	*code = strings.ReplaceAll(*code, template.PH_BR_HTTP_HANDLER_FUNCTIONS, brHandlerFuncsBuf.String())
}

// generateBRFilterDoc 生成BR关系中对方表的筛选字段文档
func generateBRFilterDoc(table *model.Table) string {
	var buf strings.Builder
	hasFilterCol := false
	for _, col := range table.Columns {
		if col.IsFilter && !col.IsHidden {
			hasFilterCol = true
			break
		}
	}
	if !hasFilterCol {
		return ""
	}

	for _, col := range table.Columns {
		if !col.IsFilter || col.IsHidden {
			continue
		}

		gotype := helper.GetGoTypeForHandler(col)
		gotype = strings.TrimPrefix(gotype, "*")
		buf.WriteString(fmt.Sprintf("\n// @Param		%s			query		%s	false	\"%s\"",
			col.Name,
			gotype,
			helper.GetCommentForHandler(col),
		))
	}
	return buf.String()
}

// generateBRFilterAssign 生成BR关系中筛选条件的赋值代码
func generateBRFilterAssign(table *model.Table) string {
	var buf strings.Builder
	hasFilterCol := false
	for _, col := range table.Columns {
		if col.IsFilter && !col.IsHidden {
			hasFilterCol = true
			break
		}
	}
	if !hasFilterCol {
		return ""
	}

	buf.WriteString(fmt.Sprintf("\n\t\tFilter: &biz.%sFilterOption{", helper.GetStructName(table.Name)))
	for _, col := range table.Columns {
		if !col.IsFilter || col.IsHidden {
			continue
		}
		buf.WriteString(fmt.Sprintf("\n\t\t\t%s: req.%s,",
			helper.GetTableColName(col.Name),
			helper.GetStructName(col.Name),
		))
	}
	buf.WriteString("\n\t\t},")
	return buf.String()
}

// generateBROrderDoc 生成BR关系中对方表的排序字段文档
func generateBROrderDoc(table *model.Table) string {
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

	orderCols := make([]string, 0)
	for _, col := range table.Columns {
		if !col.IsOrder || col.IsHidden {
			continue
		}
		orderCols = append(orderCols, col.Name)
	}
	buf.WriteString(fmt.Sprintf("\n// @Param		order_by		query		string	false	\"排序字段,可选:%s\"",
		strings.Join(orderCols, "|"),
	))
	buf.WriteString("\n// @Param		order_type		query		string	false	\"排序类型,默认desc\"")
	return buf.String()
}

// generateBRFilterPrepare 生成BR关系中筛选条件的准备赋值代码
func generateBRFilterPrepare(table *model.Table) string {
	hasFilterCol := false
	for _, col := range table.Columns {
		if col.IsFilter && !col.IsHidden {
			hasFilterCol = true
			break
		}
	}
	if !hasFilterCol {
		return ""
	}

	// 这里可以添加准备逻辑，目前保持简单
	return ""
}

// generateBROrderAssign 生成BR关系中排序条件的赋值代码
func generateBROrderAssign(table *model.Table) string {
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

	buf.WriteString("\n\t\tOrder: &option.OrderOption{")
	buf.WriteString("\n\t\t\tOrderBy:   req.OrderBy,")
	buf.WriteString("\n\t\t\tOrderType: req.OrderType,")
	buf.WriteString("\n\t\t},")
	return buf.String()
}

// genBRGRPCHandlerFunctions 为DATA表生成BR关系的gRPC handler函数
func genBRGRPCHandlerFunctions(code *string, mainTable *model.Table) {
	// 构建表名到表的映射
	tableNameToTable := make(map[string]*model.Table)
	for _, t := range mainTable.Database.Tables {
		tableNameToTable[t.Name] = t
	}

	// 获取当前表的所有BR表关系
	brTables := helper.GetMainTableBRs(mainTable, mainTable.Database.Tables)

	if len(brTables) == 0 {
		*code = strings.ReplaceAll(*code, template.PH_BR_GRPC_HANDLER_FUNCTIONS, "")
		return
	}

	var brGrpcHandlerFuncsBuf strings.Builder

	// 为每个BR表生成gRPC GET handler函数
	for _, brTable := range brTables {
		// 获取对方表
		otherTable := helper.GetBROtherTable(brTable, mainTable, tableNameToTable)
		if otherTable == nil {
			continue
		}

		// 生成gRPC handler函数
		getFunc := template.TplGRPCBRHandlerGet

		// 替换当前表相关的占位符
		getFunc = strings.ReplaceAll(getFunc, template.PH_TABLE_NAME_STRUCT, helper.GetStructName(mainTable.Name))
		getFunc = strings.ReplaceAll(getFunc, template.PH_TABLE_COMMENT, mainTable.Comment)

		// 替换对方表相关的占位符
		getFunc = strings.ReplaceAll(getFunc, template.PH_OTHER_TABLE_NAME_STRUCT, helper.GetStructName(otherTable.Name))
		getFunc = strings.ReplaceAll(getFunc, template.PH_OTHER_TABLE_COMMENT, otherTable.Comment)

		// 生成gRPC筛选条件赋值
		otherFilterAssignGrpc := generateBRFilterAssignGRPC(otherTable)
		getFunc = strings.ReplaceAll(getFunc, template.PH_OTHER_FILTER_ASSIGN_TO_OPTION_GRPC, otherFilterAssignGrpc)

		// 生成gRPC排序条件赋值
		otherOrderAssignGrpc := generateBROrderAssignGRPC(otherTable)
		getFunc = strings.ReplaceAll(getFunc, template.PH_OTHER_ASSIGN_ORDER_TO_OPTION_GRPC, otherOrderAssignGrpc)

		brGrpcHandlerFuncsBuf.WriteString(getFunc)
	}

	*code = strings.ReplaceAll(*code, template.PH_BR_GRPC_HANDLER_FUNCTIONS, brGrpcHandlerFuncsBuf.String())
}

// generateBRFilterAssignGRPC 生成BR关系中gRPC筛选条件的赋值代码
func generateBRFilterAssignGRPC(table *model.Table) string {
	var buf strings.Builder
	hasFilterCol := false
	for _, col := range table.Columns {
		if col.IsFilter && !col.IsHidden {
			hasFilterCol = true
			break
		}
	}
	if !hasFilterCol {
		return ""
	}

	buf.WriteString("\n\t\tFilter: &biz.")
	buf.WriteString(helper.GetStructName(table.Name))
	buf.WriteString("FilterOption{")
	for _, col := range table.Columns {
		if !col.IsFilter || col.IsHidden {
			continue
		}
		buf.WriteString(fmt.Sprintf("\n\t\t\t%s: req.%s,",
			helper.GetTableColName(col.Name),
			helper.GetStructName(col.Name),
		))
	}
	buf.WriteString("\n\t\t},")
	return buf.String()
}

// generateBROrderAssignGRPC 生成BR关系中gRPC排序条件的赋值代码
func generateBROrderAssignGRPC(table *model.Table) string {
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

	buf.WriteString("\n\t\tOrder: &option.OrderOption{")
	buf.WriteString("\n\t\t\tOrderBy:   req.OrderBy,")
	buf.WriteString("\n\t\t\tOrderType: req.OrderType,")
	buf.WriteString("\n\t\t},")
	return buf.String()
}
