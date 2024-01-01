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
	} else if table.Type == model.TableType_META {
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_GET_LIST, template.TplGRPCHandlerGetList)
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_UPDATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_DELETE, "")

		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_GET_LIST, template.TplHTTPHandlerGetList)
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_UPDATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_DELETE, "")
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_GET_LIST, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_UPDATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_GRPC_HANDLER_DELETE, "")

		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_GET_LIST, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_UPDATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_HTTP_HANDLER_DELETE, "")
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
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_BO_TO_VO, bufPrepare.String())
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
