package repogen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"meta-egg/internal/domain/helper"
	"meta-egg/internal/domain/repo_generator/template"
	"meta-egg/internal/model"
	"meta-egg/internal/repo"

	log "github.com/sirupsen/logrus"

	jgstr "github.com/Jinglever/go-string"
)

// relativeDir2NeedConfirm: key是相对于工程根目录的路径，值是布尔值，代表是否需要确认；注意，不会递归查看子目录
func Generate(codeDir string, project *model.Project, dbOper repo.DBOperator) (relativeDir2NeedConfirm map[string]bool, err error) {
	relativeDir2NeedConfirm = map[string]bool{
		filepath.Join("gen", "repo"):                false,
		filepath.Join("internal", "repo", "option"): true,
		filepath.Join("internal", "repo"):           true,
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

	// internal/repo/option/option.go
	path = filepath.Join(codeDir, "internal", "repo", "option", "option.go")
	err = generateGoFile(path, template.TplInternalOptionOption, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate internal/repo/option/option.go failed: %v", err)
		return
	}

	// gen/repo/base.go
	path = filepath.Join(codeDir, "gen", "repo", "base.go")
	err = generateGoFile(path, template.TplGenRepoBase, project, helper.AddHeaderNoEdit)
	if err != nil {
		log.Errorf("generate gen/repo/base.go failed: %v", err)
		return
	}

	// internal/repo/base.go
	path = filepath.Join(codeDir, "internal", "repo", "base.go")
	err = generateGoFile(path, template.TplInternalRepoBase, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate internal/repo/base.go failed: %v", err)
		return
	}

	if project.Database == nil || len(project.Database.Tables) == 0 {
		return
	}

	// 按照表生成文件
	for _, table := range project.Database.Tables {
		// gen
		tpl := getGenTableTplByTableType(table.Type)
		err = generateGoFileForTable(filepath.Join(codeDir, "gen", "repo", table.Name+".go"),
			tpl, table, helper.AddHeaderNoEdit, dbOper)
		if err != nil {
			log.Errorf("generate gen/repo/%s.go failed: %v", table.Name, err)
			return
		}

		// internal
		tpl = getInternalTableTplByTableType(table.Type)
		err = generateGoFileForTable(filepath.Join(codeDir, "internal", "repo", table.Name+".go"),
			tpl, table, helper.AddHeaderCanEdit, nil)
		if err != nil {
			log.Errorf("generate internal/repo/%s.go failed: %v", table.Name, err)
			return
		}

		if table.Type == model.TableType_DATA || table.Type == model.TableType_META {
			// internal/repo/option
			err = generateGoFileForTable(filepath.Join(codeDir, "internal", "repo", "option", table.Name+".go"),
				template.TplInternalOptionTable, table, helper.AddHeaderCanEdit, nil)
			if err != nil {
				log.Errorf("generate internal/repo/option/%s.go failed: %v", table.Name, err)
				return
			}
		}
	}
	return
}

func getGenTableTplByTableType(tableType model.TableType) string {
	switch tableType {
	case model.TableType_DATA:
		return template.TplGenRepoDataTable
	case model.TableType_BR:
		return template.TplGenRepoDataTable
	case model.TableType_RL:
		return template.TplGenRepoDataTable
	case model.TableType_META:
		return template.TplGenRepoMetaTable
	default:
		return ""
	}
}

func getInternalTableTplByTableType(tableType model.TableType) string {
	switch tableType {
	case model.TableType_DATA:
		return template.TplInternalRepoTableData
	case model.TableType_BR:
		return template.TplInternalRepoTableMeta
	case model.TableType_RL:
		return template.TplInternalRepoTableMeta
	case model.TableType_META:
		return template.TplInternalRepoTableMeta
	default:
		return ""
	}
}

func generateGoFileForTable(path string, tpl string, table *model.Table, addHeader func(s string) string, dbOper repo.DBOperator) error {
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
	err = replaceTplForTable(&code, table, dbOper)
	if err != nil {
		log.Errorf("replace template for table failed: %v", err)
		return err
	}

	// go format
	formatted, err := jgstr.FormatGo([]byte(code))
	if err != nil {
		log.Errorf("format source failed: %v\n%s", err, code)
		return err
	}
	_, _ = f.Write(formatted)
	return nil
}

func replaceTplForTable(code *string, table *model.Table, dbOper repo.DBOperator) error {
	eCols := helper.GetExtTypeCols(table)
	genSetupMEForCreate(code, table, eCols)
	genSetupMEForUpdate(code, table, eCols)
	genSetupMEForDelete(code, table, eCols)
	genOrderByList(code, table)
	genFilterColList(code, table)
	genFilterGetRepoOptions(code, table)
	if table.Type == model.TableType_META {
		metaRecords, err := helper.GetMetaRecords(table, eCols, dbOper)
		if err != nil {
			log.Errorf("get meta records failed: %v", err)
			return err
		}
		genCaseMetaIDToSemantic(code, table, eCols, metaRecords)
	}

	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, table.Database.Project.GoModule)
	*code = strings.ReplaceAll(*code, template.PH_TABLE_NAME, table.Name)
	*code = strings.ReplaceAll(*code, template.PH_STRUCT_TABLE_NAME, helper.GetStructName(table.Name))
	*code = strings.ReplaceAll(*code, template.PH_COL_ID, helper.GetTableColName(table.PrimaryColumn.Name))

	if eCols.UpdatedBy != nil {
		*code = strings.ReplaceAll(*code, template.PH_COL_UPDATED_BY, helper.GetTableColName(eCols.UpdatedBy.Name))
	}
	if eCols.DeletedBy != nil {
		*code = strings.ReplaceAll(*code, template.PH_COL_DELETED_BY, helper.GetTableColName(eCols.DeletedBy.Name))
	}
	if eCols.DeletedAt != nil {
		*code = strings.ReplaceAll(*code, template.PH_COL_DELETED_AT, helper.GetTableColName(eCols.DeletedAt.Name))
	}

	return nil
}

func genCaseMetaIDToSemantic(code *string, table *model.Table, eCols *helper.ExtTypeCols, metaRecords []map[string]interface{}) {
	if len(metaRecords) == 0 {
		*code = strings.ReplaceAll(*code, template.PH_CASE_META_ID_TO_SEMANTIC, "")
	} else {
		var buf strings.Builder
		// case meta id to semantic
		for _, record := range metaRecords {
			buf.WriteString(fmt.Sprintf("\ncase model.Meta%s%s:\n		return \"%s\"",
				helper.GetStructName(table.Name),
				helper.GetTableColName(record[eCols.Semantic.Name].(string)),
				strings.TrimSpace(record[eCols.Semantic.Name].(string))))
		}
		*code = strings.ReplaceAll(*code, template.PH_CASE_META_ID_TO_SEMANTIC, buf.String())
	}
}

func genSetupMEForCreate(code *string, table *model.Table, eCols *helper.ExtTypeCols) {
	if eCols.CreatedBy != nil || eCols.UpdatedBy != nil {
		*code = strings.ReplaceAll(*code, template.PH_TPL_SETUP_ME_FOR_CREATE, template.TplSetupMEForCreate)
		*code = strings.ReplaceAll(*code, template.PH_TPL_SETUP_ME_FOR_CREATE_BATCH, template.TplSetupMEForCreateBatch)
		*code = strings.ReplaceAll(*code, template.PH_SET_ME_FOR_CREATE, genSetMEForCreate(eCols))
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_SETUP_ME_FOR_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_SETUP_ME_FOR_CREATE_BATCH, "")
	}
}

func genSetupMEForUpdate(code *string, table *model.Table, eCols *helper.ExtTypeCols) {
	if eCols.UpdatedBy != nil {
		*code = strings.ReplaceAll(*code, template.PH_TPL_SETUP_ME_FOR_UPDATE, template.TplSetupMEForUpdate)
		*code = strings.ReplaceAll(*code, template.PH_COL_UPDATED_BY, helper.GetTableColName(eCols.UpdatedBy.Name))
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_SETUP_ME_FOR_UPDATE, "")
	}
}

func genSetupMEForDelete(code *string, table *model.Table, eCols *helper.ExtTypeCols) {
	if eCols.DeletedAt != nil && eCols.DeletedBy != nil {
		*code = strings.ReplaceAll(*code, template.PH_TPL_DELETE, template.TplSoftDeleteWithDeletedBy)
		*code = strings.ReplaceAll(*code, template.PH_COL_DELETED_BY, helper.GetTableColName(eCols.DeletedBy.Name))
		*code = strings.ReplaceAll(*code, template.PH_COL_DELETED_AT, helper.GetTableColName(eCols.DeletedAt.Name))
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_DELETE, template.TplDefaultDelete)
	}
}

func genSetMEForCreate(eCols *helper.ExtTypeCols) string {
	var buf strings.Builder
	if (eCols.CreatedBy != nil && !eCols.CreatedBy.IsRequired) ||
		(eCols.UpdatedBy != nil && !eCols.UpdatedBy.IsRequired) {
		buf.WriteString("meID := me.ID\n")
	}
	if eCols.CreatedBy != nil {
		if eCols.CreatedBy.IsRequired {
			buf.WriteString(fmt.Sprintf("m.%s = me.ID", helper.GetTableColName(eCols.CreatedBy.Name)))
		} else {
			buf.WriteString(fmt.Sprintf("m.%s = &meID", helper.GetTableColName(eCols.CreatedBy.Name)))
		}
	}
	if eCols.UpdatedBy != nil {
		if eCols.CreatedBy != nil {
			buf.WriteString("\n")
		}
		if eCols.UpdatedBy.IsRequired {
			buf.WriteString(fmt.Sprintf("m.%s = me.ID", helper.GetTableColName(eCols.UpdatedBy.Name)))
		} else {
			buf.WriteString(fmt.Sprintf("m.%s = &meID", helper.GetTableColName(eCols.UpdatedBy.Name)))
		}
	}
	return buf.String()
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
	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, project.GoModule)
	genNewRepoFuncListInProviderSet(code, project)
	genNewRepoFuncListInMockProviderSet(code, project)
}

func genNewRepoFuncListInProviderSet(code *string, project *model.Project) {
	var buf strings.Builder
	if project.Database != nil {
		for _, table := range project.Database.Tables {
			buf.WriteString(fmt.Sprintf("New%sRepo,\n", helper.GetStructName(table.Name)))
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_NEW_REPO_FUNC_LIST_IN_PROVIDER_SET, buf.String())
}

func genNewRepoFuncListInMockProviderSet(code *string, project *model.Project) {
	var buf strings.Builder
	if project.Database != nil {
		for _, table := range project.Database.Tables {
			buf.WriteString(fmt.Sprintf("// mock.NewMock%sRepo,\n", helper.GetStructName(table.Name)))
			buf.WriteString(fmt.Sprintf("// wire.Bind(new(%sRepo), new(*mock.Mock%sRepo)),\n",
				helper.GetStructName(table.Name),
				helper.GetStructName(table.Name)))
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_NEW_REPO_FUNC_LIST_IN_MOCK_PROVIDER_SET, buf.String())
}

// PH_ORDER_BY_LIST
// like:
// model.ColUserID,
func genOrderByList(code *string, table *model.Table) {
	var buf strings.Builder
	for _, col := range table.Columns {
		if !col.IsOrder {
			continue
		}
		buf.WriteString(fmt.Sprintf("model.Col%s%s,\n",
			helper.GetStructName(table.Name),
			helper.GetTableColName(col.Name),
		))
	}
	*code = strings.ReplaceAll(*code, template.PH_ORDER_BY_LIST, buf.String())
}

// PH_FILTER_COL_LIST
func genFilterColList(code *string, table *model.Table) {
	var buf bytes.Buffer
	for _, col := range table.Columns {
		if col.IsFilter {
			gotype, err := helper.GetGoType(col)
			if err != nil {
				log.Fatalf("get go type failed: %v", err)
				return
			}
			if !helper.IsGoTypeNullable(gotype) {
				gotype = "*" + gotype
			}
			buf.WriteString(fmt.Sprintf("%s %s // %s\n",
				helper.GetTableColName(col.Name),
				gotype,
				col.Comment,
			))
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_FILTER_COL_LIST, buf.String())
}

func genFilterGetRepoOptions(code *string, table *model.Table) {
	var buf bytes.Buffer
	for _, col := range table.Columns {
		if col.IsFilter {
			buf.WriteString(fmt.Sprintf("if o.%s != nil {\n", helper.GetTableColName(col.Name)))

			gotype, err := helper.GetGoType(col)
			if err != nil {
				log.Fatalf("get go type failed: %v", err)
				return
			}
			star := ""
			if !helper.IsGoTypeNullable(gotype) {
				star = "*"
			}

			buf.WriteString(fmt.Sprintf("\tops = append(ops, gormx.Where(model.Col%s%s+\" = ?\", %so.%s))\n",
				helper.GetStructName(table.Name),
				helper.GetTableColName(col.Name),
				star,
				helper.GetTableColName(col.Name),
			))
			buf.WriteString("\t\t}\n")
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_FILTER_GET_REPO_OPTIONS, buf.String())
}
