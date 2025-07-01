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
		// Skip RL tables as they don't need independent repo files
		// RL table operations are integrated into their main table's repo
		if table.Type == model.TableType_RL {
			continue
		}

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

	// Generate RL table related code for DATA tables
	if table.Type == model.TableType_DATA {
		genRLMethods(code, table)
	}

	if table.Type == model.TableType_META {
		metaRecords, err := helper.GetMetaRecords(table, eCols, dbOper)
		if err != nil {
			log.Errorf("get meta records failed: %v", err)
			return err
		}
		genCaseMetaSemanticToID(code, table, eCols, metaRecords)
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
		if eCols.DeletedAt.ExtType == model.ColumnExtType_TIME_DELETE {
			*code = strings.ReplaceAll(*code, template.PH_VAL_DELETED_AT, "time.Now()")
		} else if eCols.DeletedAt.ExtType == model.ColumnExtType_TIME_DELETE2 {
			*code = strings.ReplaceAll(*code, template.PH_VAL_DELETED_AT, "time.Now().Unix()")
		}
	}

	return nil
}

func genCaseMetaSemanticToID(code *string, table *model.Table, eCols *helper.ExtTypeCols, metaRecords []map[string]interface{}) {
	if len(metaRecords) == 0 {
		*code = strings.ReplaceAll(*code, template.PH_CASE_META_SEMANTIC_TO_ID, "")
	} else {
		var buf strings.Builder
		// case meta id to semantic
		for _, record := range metaRecords {
			buf.WriteString(fmt.Sprintf("\ncase \"%s\":\n		return model.Meta%s%s",
				strings.TrimSpace(record[eCols.Semantic.Name].(string)),
				helper.GetStructName(table.Name),
				helper.GetTableColName(record[eCols.Semantic.Name].(string)),
			))
		}
		*code = strings.ReplaceAll(*code, template.PH_CASE_META_SEMANTIC_TO_ID, buf.String())
	}
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

func genRLMethods(code *string, table *model.Table) {
	// 构建表名到表的映射
	tableNameToTable := make(map[string]*model.Table)
	for _, t := range table.Database.Tables {
		tableNameToTable[t.Name] = t
	}

	// 获取当前主表的所有RL表
	rlTables := helper.GetMainTableRLs(table, table.Database.Tables)

	if len(rlTables) == 0 {
		// 如果没有RL表，清空占位符
		*code = strings.ReplaceAll(*code, template.PH_RL_METHODS_INTERFACE, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_METHODS_IMPLEMENTATION, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_LIST_PRELOAD, "")
		*code = strings.ReplaceAll(*code, template.PH_RL_DETAIL_PRELOAD, "")
		return
	}

	var interfaceBuf, implBuf, listPreloadBuf, detailPreloadBuf strings.Builder

	// 生成预加载逻辑
	for _, rlTable := range rlTables {
		rlStructName := helper.GetStructName(rlTable.Name)
		rlFieldName := rlStructName + "s" // 复数形式

		// 详情查询总是预加载所有RL表
		detailPreloadBuf.WriteString(fmt.Sprintf("\ttx = tx.Preload(\"%s\")\n", rlFieldName))

		// 列表查询只预加载有list字段的RL表
		if helper.ShouldIncludeRLInList(rlTable) {
			listPreloadBuf.WriteString(fmt.Sprintf("\ttx = tx.Preload(\"%s\")\n", rlFieldName))
		}
	}

	// 生成RL表操作方法
	for _, rlTable := range rlTables {
		rlStructName := helper.GetStructName(rlTable.Name)
		mainTableStructName := helper.GetStructName(table.Name)
		rlTablePrimaryColName := helper.GetTableColName(rlTable.PrimaryColumn.Name)

		// 找到指向主表的外键字段名
		var mainFKFieldName string

		// 找到指向主表的外键字段
		for _, column := range rlTable.Columns {
			for _, fk := range column.ForeignKeys {
				if fk.Table == table.Name {
					// 如果有明确标记为主外键，或者只有一个外键指向主表
					if fk.IsMain {
						mainFKFieldName = helper.GetTableColName(column.Name)
						break
					} else if mainFKFieldName == "" {
						// 如果还没找到，先记录这个
						mainFKFieldName = helper.GetTableColName(column.Name)
					}
				}
			}
			if mainFKFieldName != "" {
				break
			}
		}

		if mainFKFieldName == "" {
			continue // 跳过无法识别主外键的RL表
		}

		// 生成接口方法
		interfaceBuf.WriteString(fmt.Sprintf(`
	// %s related methods
	Add%s(ctx context.Context, mainId uint64, rl *model.%s) error
	Remove%s(ctx context.Context, mainId uint64, rlId uint64) error
	RemoveAll%s(ctx context.Context, mainId uint64) error
	GetAll%s(ctx context.Context, mainId uint64) ([]*model.%s, error)`,
			rlStructName, rlStructName, rlStructName, rlStructName, rlStructName, rlStructName, rlStructName))

		// 生成实现方法 - 分别生成每个方法避免参数混乱

		// Add方法
		implBuf.WriteString(fmt.Sprintf(`
// Add%s adds a %s record for the main table
func (s *%sRepoImpl) Add%s(ctx context.Context, mainId uint64, rl *model.%s) error {
	if rl == nil {
		return fmt.Errorf("rl cannot be nil")
	}
	rl.%s = mainId
	%s
	return s.GetTX(ctx).Create(rl).Error
}`,
			rlStructName, rlTable.Name, mainTableStructName, rlStructName, rlStructName, mainFKFieldName,
			generateRLSetMEForCreate(rlTable)))

		// Remove方法
		implBuf.WriteString(fmt.Sprintf(`
// Remove%s removes a %s record by RL ID
func (s *%sRepoImpl) Remove%s(ctx context.Context, mainId uint64, rlId uint64) error {
	tx := s.GetTX(ctx).Where(model.Col%s%s+" = ? AND "+model.Col%s%s+" = ?", mainId, rlId)
	%s
	return nil
}`,
			rlStructName, rlTable.Name, mainTableStructName, rlStructName,
			rlStructName, mainFKFieldName, rlStructName, rlTablePrimaryColName,
			generateRLDeleteLogic(rlTable, rlStructName)))

		// RemoveAll方法
		implBuf.WriteString(fmt.Sprintf(`
// RemoveAll%s removes all %s records for the main table
func (s *%sRepoImpl) RemoveAll%s(ctx context.Context, mainId uint64) error {
	tx := s.GetTX(ctx).Where(model.Col%s%s+" = ?", mainId)
	%s
	return nil
}`,
			rlStructName, rlTable.Name, mainTableStructName, rlStructName,
			rlStructName, mainFKFieldName,
			generateRLBatchDeleteLogic(rlTable, rlStructName)))

		// Get方法
		implBuf.WriteString(fmt.Sprintf(`
// Get%ss retrieves all %s records for the main table
func (s *%sRepoImpl) GetAll%s(ctx context.Context, mainId uint64) ([]*model.%s, error) {
	var rls []*model.%s
	err := s.GetTX(ctx).Where(model.Col%s%s+" = ?", mainId).Find(&rls).Error
	return rls, err
}`,
			rlStructName, rlTable.Name, mainTableStructName, rlStructName, rlStructName, rlStructName,
			rlStructName, mainFKFieldName))
	}

	// 替换占位符
	*code = strings.ReplaceAll(*code, template.PH_RL_METHODS_INTERFACE, interfaceBuf.String())
	*code = strings.ReplaceAll(*code, template.PH_RL_METHODS_IMPLEMENTATION, implBuf.String())
	*code = strings.ReplaceAll(*code, template.PH_RL_LIST_PRELOAD, listPreloadBuf.String())
	*code = strings.ReplaceAll(*code, template.PH_RL_DETAIL_PRELOAD, detailPreloadBuf.String())
}

func generateRLSetMEForCreate(rlTable *model.Table) string {
	eCols := helper.GetExtTypeCols(rlTable)
	if eCols.CreatedBy == nil && eCols.UpdatedBy == nil {
		return ""
	}

	var buf strings.Builder
	buf.WriteString("if me, ok := contexts.GetME(ctx); ok {\n\t\t")

	if (eCols.CreatedBy != nil && !eCols.CreatedBy.IsRequired) ||
		(eCols.UpdatedBy != nil && !eCols.UpdatedBy.IsRequired) {
		buf.WriteString("meID := me.ID\n\t\t")
	}

	if eCols.CreatedBy != nil {
		if eCols.CreatedBy.IsRequired {
			buf.WriteString(fmt.Sprintf("rl.%s = me.ID", helper.GetTableColName(eCols.CreatedBy.Name)))
		} else {
			buf.WriteString(fmt.Sprintf("rl.%s = &meID", helper.GetTableColName(eCols.CreatedBy.Name)))
		}
	}

	if eCols.UpdatedBy != nil {
		if eCols.CreatedBy != nil {
			buf.WriteString("\n\t\t")
		}
		if eCols.UpdatedBy.IsRequired {
			buf.WriteString(fmt.Sprintf("rl.%s = me.ID", helper.GetTableColName(eCols.UpdatedBy.Name)))
		} else {
			buf.WriteString(fmt.Sprintf("rl.%s = &meID", helper.GetTableColName(eCols.UpdatedBy.Name)))
		}
	}

	buf.WriteString("\n\t}")
	return buf.String()
}

func generateRLDeleteLogic(rlTable *model.Table, rlStructName string) string {
	eCols := helper.GetExtTypeCols(rlTable)
	if eCols.DeletedAt != nil && eCols.DeletedBy != nil {
		// 软删除：有删除者和删除时间字段
		var deletedAtValue string
		if eCols.DeletedAt.ExtType == model.ColumnExtType_TIME_DELETE {
			deletedAtValue = "time.Now()"
		} else if eCols.DeletedAt.ExtType == model.ColumnExtType_TIME_DELETE2 {
			deletedAtValue = "time.Now().Unix()"
		}

		return fmt.Sprintf(`var result *gorm.DB
	if me, ok := contexts.GetME(ctx); ok {
		result = tx.UpdateColumns(map[string]interface{}{
			model.Col%s%s: &(me.ID),
			model.Col%s%s: %s,
		})
	} else {
		result = tx.Delete(&model.%s{})
	}
	if result.Error != nil {
		return result.Error
	}`,
			rlStructName, helper.GetTableColName(eCols.DeletedBy.Name),
			rlStructName, helper.GetTableColName(eCols.DeletedAt.Name), deletedAtValue,
			rlStructName)
	} else {
		// 硬删除：没有软删除字段
		return fmt.Sprintf(`result := tx.Delete(&model.%s{})
	if result.Error != nil {
		return result.Error
	}`, rlStructName)
	}
}

func generateRLBatchDeleteLogic(rlTable *model.Table, rlStructName string) string {
	eCols := helper.GetExtTypeCols(rlTable)
	if eCols.DeletedAt != nil && eCols.DeletedBy != nil {
		// 软删除：有删除者和删除时间字段
		var deletedAtValue string
		if eCols.DeletedAt.ExtType == model.ColumnExtType_TIME_DELETE {
			deletedAtValue = "time.Now()"
		} else if eCols.DeletedAt.ExtType == model.ColumnExtType_TIME_DELETE2 {
			deletedAtValue = "time.Now().Unix()"
		}

		return fmt.Sprintf(`var result *gorm.DB
	if me, ok := contexts.GetME(ctx); ok {
		result = tx.UpdateColumns(map[string]interface{}{
			model.Col%s%s: &(me.ID),
			model.Col%s%s: %s,
		})
	} else {
		result = tx.Delete(&model.%s{})
	}
	if result.Error != nil {
		return result.Error
	}`,
			rlStructName, helper.GetTableColName(eCols.DeletedBy.Name),
			rlStructName, helper.GetTableColName(eCols.DeletedAt.Name), deletedAtValue,
			rlStructName)
	} else {
		// 硬删除：没有软删除字段
		return fmt.Sprintf(`result := tx.Delete(&model.%s{})
	if result.Error != nil {
		return result.Error
	}`, rlStructName)
	}
}
