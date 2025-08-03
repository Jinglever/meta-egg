package bizgen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"meta-egg/internal/domain/biz_generator/template"
	"meta-egg/internal/domain/helper"
	"meta-egg/internal/model"

	jgstr "github.com/Jinglever/go-string"
	log "github.com/sirupsen/logrus"
)

// relativeDir2NeedConfirm: key是相对于工程根目录的路径，值是布尔值，代表是否需要确认；注意，不会递归查看子目录
func Generate(codeDir string, project *model.Project) (relativeDir2NeedConfirm map[string]bool, err error) {
	relativeDir2NeedConfirm = map[string]bool{
		filepath.Join("internal", "biz"): true,
		// filepath.Join("internal", "biz", "option"): true,
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

	// // internal/biz/option/option.go
	// path = filepath.Join(codeDir, "internal", "biz", "option", "option.go")
	// err = generateGoFile(path, template.TplOptionOption, project, helper.AddHeaderCanEdit)
	// if err != nil {
	// 	log.Errorf("generate internal/biz/option/option.go failed: %v", err)
	// 	return
	// }

	// internal/biz/base.go
	path = filepath.Join(codeDir, "internal", "biz", "base.go")
	err = generateGoFile(path, template.TplBase, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate internal/biz/base.go failed: %v", err)
		return
	}

	// internal/biz/wire_gen.go
	path = filepath.Join(codeDir, "internal", "biz", "wire_gen.go")
	err = generateGoFile(path, template.TplInternalBizWireGen, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate internal/biz/wire_gen.go failed: %v", err)
		return
	}

	// internal/biz/wire.go
	path = filepath.Join(codeDir, "internal", "biz", "wire.go")
	err = generateGoFile(path, template.TplInternalBizWire, project, helper.AddHeaderCanEdit)
	if err != nil {
		log.Errorf("generate internal/biz/wire.go failed: %v", err)
		return
	}

	if project.Database != nil {
		for _, table := range project.Database.Tables {
			if !table.HasHandler { // 由于biz里的代码都是面向handler的, 所以handler不开启的情况下, biz也不要生成
				continue
			}
			// internal/biz/<table>.go
			path = filepath.Join(codeDir, "internal", "biz", helper.GetDirName(table.Name)+".go")
			err = generateGoFileForTable(path, template.TplTable, table, helper.AddHeaderCanEdit)
			if err != nil {
				log.Errorf("generate internal/biz/%s.go failed: %v",
					helper.GetDirName(table.Name), err)
				return
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
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_CREATE, template.TplFuncCreate)
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_GET_LIST, template.TplFuncGetList)
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_UPDATE, template.TplFuncUpdate)
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_DELETE, template.TplFuncDelete)
		// Generate RL table related code for DATA tables
		genRLBizMethods(code, table)
		// Generate BR table related code for DATA tables
		genBRBizMethods(code, table)
	} else if table.Type == model.TableType_META {
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_GET_LIST, template.TplFuncGetList)
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_UPDATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_DELETE, "")
		// Clear RL placeholders for non-DATA tables
		clearRLPlaceholders(code)
		// Clear BR placeholders for non-DATA tables
		clearBRPlaceholders(code)
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_GET_LIST, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_UPDATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_DELETE, "")
		// Clear RL placeholders for non-DATA tables
		clearRLPlaceholders(code)
		// Clear BR placeholders for non-DATA tables
		clearBRPlaceholders(code)
	}

	genErrorDuplicateKey(code, table)
	genFilterColList(code, table)
	genFilterGetRepoOptions(code, table)
	getOrderByList(code, table)
	genSetColList(code, table)
	genSetUpdateSetCVs(code, table)

	genColListInBO(code, table)
	genAssignModelToBO(code, table)
	genAssignBOToModel(code, table)
	genColListForList(code, table)
	genAssignModelForList(code, table)
	genAssignFilterToOption(code, table)
	genColListToSelectForList(code, table)

	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, table.Database.Project.GoModule)
	*code = strings.ReplaceAll(*code, template.PH_TABLE_NAME, table.Name)
	*code = strings.ReplaceAll(*code, template.PH_STRUCT_TABLE_NAME, helper.GetStructName(table.Name))
	*code = strings.ReplaceAll(*code, template.PH_TABLE_NAME_VAR, helper.GetVarName(table.Name))
}

func genErrorDuplicateKey(code *string, table *model.Table) {
	hasUniqueKey := false
	if table.Type == model.TableType_DATA {
		for _, col := range table.Columns {
			if col.IsUnique {
				hasUniqueKey = true
				break
			}
		}
		if len(table.Unique) > 0 {
			hasUniqueKey = true
		}
	}
	if hasUniqueKey {
		*code = strings.ReplaceAll(*code, template.PH_CREATE_ERROR_DUPLICATE_KEY, template.TplCreateErrorDuplicateKey)
		*code = strings.ReplaceAll(*code, template.PH_UPDATE_ERROR_DUPLICATE_KEY, template.TplUpdateErrorDuplicateKey)
	} else {
		*code = strings.ReplaceAll(*code, template.PH_CREATE_ERROR_DUPLICATE_KEY, "")
		*code = strings.ReplaceAll(*code, template.PH_UPDATE_ERROR_DUPLICATE_KEY, "")
	}
}

func genAssignRepoList(code *string, project *model.Project) {
	if project.Database == nil || len(project.Database.Tables) == 0 {
		*code = strings.ReplaceAll(*code, template.PH_ASSIGN_REPO_LIST, "")
		return
	}
	var buf bytes.Buffer
	for _, table := range project.Database.Tables {
		// Skip RL tables as they don't have independent repo implementations
		// RL table operations are integrated into their main table's repo
		if table.Type == model.TableType_RL {
			continue
		}
		buf.WriteString(fmt.Sprintf("%sRepo: %sRepo,\n",
			helper.GetStructName(table.Name),
			helper.GetVarName(table.Name)))
	}
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_REPO_LIST, buf.String())
}

func genRepoListInArg(code *string, project *model.Project) {
	if project.Database == nil || len(project.Database.Tables) == 0 {
		*code = strings.ReplaceAll(*code, template.PH_REPO_LIST_IN_ARG, "")
		return
	}
	var buf bytes.Buffer
	for _, table := range project.Database.Tables {
		// Skip RL tables as they don't have independent repo implementations
		// RL table operations are integrated into their main table's repo
		if table.Type == model.TableType_RL {
			continue
		}
		buf.WriteString(fmt.Sprintf("%sRepo repo.%sRepo,\n",
			helper.GetVarName(table.Name),
			helper.GetStructName(table.Name)))
	}
	*code = strings.ReplaceAll(*code, template.PH_REPO_LIST_IN_ARG, buf.String())
}

func genRepoListInStruct(code *string, project *model.Project) {
	if project.Database == nil || len(project.Database.Tables) == 0 {
		*code = strings.ReplaceAll(*code, template.PH_REPO_LIST_IN_STRUCT, "")
		return
	}
	var buf bytes.Buffer
	for _, table := range project.Database.Tables {
		// Skip RL tables as they don't have independent repo implementations
		// RL table operations are integrated into their main table's repo
		if table.Type == model.TableType_RL {
			continue
		}
		buf.WriteString(fmt.Sprintf("%sRepo repo.%sRepo\n",
			helper.GetStructName(table.Name),
			helper.GetStructName(table.Name)))
	}
	*code = strings.ReplaceAll(*code, template.PH_REPO_LIST_IN_STRUCT, buf.String())
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
	genRepoListInStruct(code, project)
	genRepoListInArg(code, project)
	genAssignRepoList(code, project)

	*code = strings.ReplaceAll(*code, template.PH_GO_MODULE, project.GoModule)

	if project.Database == nil || len(project.Database.Tables) == 0 {
		*code = strings.ReplaceAll(*code, template.PH_COMMENT_REPO, "//")
	} else {
		*code = strings.ReplaceAll(*code, template.PH_COMMENT_REPO, "")
	}
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

// PH_FILTER_GET_REPO_OPTIONS
// like:
//
//	if o.Gender != nil {
//		ops = append(ops, gormx.Where(model.ColUserGender+" = ?", *o.Gender))
//	}
func genFilterGetRepoOptions(code *string, table *model.Table) {
	var buf bytes.Buffer
	for _, col := range table.Columns {
		if col.IsFilter {
			buf.WriteString(fmt.Sprintf("if o.%s != nil {\n", helper.GetTableColName(col.Name)))
			buf.WriteString(fmt.Sprintf("\tops = append(ops, gormx.Where(model.Col%s%s+\" = ?\", *o.%s))\n",
				helper.GetStructName(table.Name),
				helper.GetTableColName(col.Name),
				helper.GetTableColName(col.Name),
			))
			buf.WriteString("\t\t}\n")
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_FILTER_GET_REPO_OPTIONS, buf.String())
}

// PH_ORDER_BY_LIST
// like:
// model.ColUserID,
func getOrderByList(code *string, table *model.Table) {
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

// PH_SET_COL_LIST
// like:
// Name   *string
// Gender *uint64
func genSetColList(code *string, table *model.Table) {
	var buf bytes.Buffer
	for _, col := range table.Columns {
		if !col.IsAlterable {
			continue
		}
		gotype, err := helper.GetGoType(col)
		if err != nil {
			log.Fatalf("get go type failed: %v", err)
			return
		}
		if !helper.IsGoTypeNullable(gotype) {
			gotype = "*" + gotype
		}
		buf.WriteString(fmt.Sprintf("\n%s %s",
			helper.GetTableColName(col.Name),
			gotype,
		))
	}
	*code = strings.ReplaceAll(*code, template.PH_SET_COL_LIST, buf.String())
}

// PH_SET_UPDATE_SETCVS
// like:
//
//	if opt.Set != nil {
//		if opt.Set.Name != nil {
//			setCVs[model.ColUserName] = *opt.Set.Name
//		}
//	}
func genSetUpdateSetCVs(code *string, table *model.Table) {
	var buf bytes.Buffer
	hasAlterableCol := false
	for _, col := range table.Columns {
		if !col.IsAlterable {
			continue
		}
		if !hasAlterableCol {
			hasAlterableCol = true
			buf.WriteString("if setOpt != nil {\n")
		}
		buf.WriteString(fmt.Sprintf("  if setOpt.%s != nil {\n", helper.GetTableColName(col.Name)))

		gotype, err := helper.GetGoType(col)
		if err != nil {
			log.Fatalf("get go type failed: %v", err)
			return
		}
		star := ""
		if !helper.IsGoTypeNullable(gotype) {
			star = "*"
		}

		buf.WriteString(fmt.Sprintf("    setCVs[model.Col%s%s] = %ssetOpt.%s\n",
			helper.GetStructName(table.Name),
			helper.GetTableColName(col.Name),
			star,
			helper.GetTableColName(col.Name),
		))
		buf.WriteString("  }\n")
	}
	if hasAlterableCol {
		buf.WriteString("}\n")
	}
	*code = strings.ReplaceAll(*code, template.PH_SET_UPDATE_SETCVS, buf.String())
}

func genColListInBO(code *string, table *model.Table) {
	var (
		buf    strings.Builder
		goType string
		err    error
	)
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}

		goType, err = helper.GetGoType(col)
		if err != nil {
			log.Fatalf("fail to get to type: %v", err)
		}
		if !col.IsRequired && !helper.IsGoTypeNullable(goType) {
			goType = "*" + goType
		}
		comment := col.Comment
		buf.WriteString(fmt.Sprintf("\n%s %s `json:\"%s\"` // %s",
			helper.GetTableColName(col.Name),
			goType,
			helper.GetDirName(col.Name),
			comment))
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_IN_BO, buf.String())
}

func genAssignModelToBO(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}

		buf.WriteString(fmt.Sprintf("\n%s: m.%s,",
			helper.GetTableColName(col.Name),
			helper.GetTableColName(col.Name),
		))
	}
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_MODEL_TO_BO, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_MODEL_TO_BO, buf.String())
}

func genAssignBOToModel(code *string, table *model.Table) {
	var buf strings.Builder
	var bufPrepare strings.Builder
	for _, col := range table.Columns {
		if !col.IsAlterable {
			continue
		}
		if col.IsHidden {
			continue
		}
		buf.WriteString(fmt.Sprintf("\n%s: obj.%s,",
			helper.GetTableColName(col.Name),
			helper.GetTableColName(col.Name),
		))
	}
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_BO_TO_MODEL, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_BO_TO_MODEL, buf.String())
}

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

		goType, err = helper.GetGoType(col)
		if err != nil {
			log.Fatalf("fail to get to type: %v", err)
		}
		if !col.IsRequired && !helper.IsGoTypeNullable(goType) {
			goType = "*" + goType
		}
		comment := col.Comment
		buf.WriteString(fmt.Sprintf("\n%s %s `json:\"%s\"` // %s",
			helper.GetTableColName(col.Name),
			goType,
			helper.GetDirName(col.Name),
			comment))
	}
	*code = strings.ReplaceAll(*code, template.PH_COL_LIST_FOR_LIST, buf.String())
}

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

		buf.WriteString(fmt.Sprintf("\n%s: ms[i].%s,",
			helper.GetTableColName(col.Name),
			helper.GetTableColName(col.Name),
		))
	}
	*code = strings.ReplaceAll(*code, template.PH_PREPARE_ASSIGN_MODEL_FOR_LIST, bufPrepare.String())
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_MODEL_FOR_LIST, buf.String())
}

func genAssignFilterToOption(code *string, table *model.Table) {
	var buf strings.Builder
	hasFilterCol := false
	for _, col := range table.Columns {
		if col.IsFilter && !col.IsHidden {
			hasFilterCol = true
			break
		}
	}
	if !hasFilterCol {
		*code = strings.ReplaceAll(*code, template.PH_ASSIGN_FILTER_TO_OPTION, "")
		return
	}
	buf.WriteString(fmt.Sprintf("\nFilter: &option.%sFilterOption{", helper.GetStructName(table.Name)))
	for _, col := range table.Columns {
		if !col.IsFilter || col.IsHidden {
			continue
		}
		buf.WriteString(fmt.Sprintf("\n%s: opt.Filter.%s,",
			helper.GetTableColName(col.Name),
			helper.GetTableColName(col.Name),
		))
	}
	buf.WriteString("\n},")
	*code = strings.ReplaceAll(*code, template.PH_ASSIGN_FILTER_TO_OPTION, buf.String())
}

// PH_COL_LIST_TO_SELECT_FOR_LIST
func genColListToSelectForList(code *string, table *model.Table) {
	var (
		buf strings.Builder
	)
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

func genRLBizMethods(code *string, table *model.Table) {
	// 构建表名到表的映射
	tableNameToTable := make(map[string]*model.Table)
	for _, t := range table.Database.Tables {
		tableNameToTable[t.Name] = t
	}

	// 获取当前主表的所有RL表
	rlTables := helper.GetMainTableRLs(table, table.Database.Tables)

	if len(rlTables) == 0 {
		clearRLPlaceholders(code)
		return
	}

	var boDefinitionsBuf, listBoDefinitionsBuf, boFieldsBuf, listBoFieldsBuf, assignModelToBOBuf, assignModelToListBOBuf, methodsBuf strings.Builder

	// 生成RL表的BO结构定义
	for _, rlTable := range rlTables {
		rlStructName := helper.GetStructName(rlTable.Name)
		boDefinitionsBuf.WriteString(fmt.Sprintf(`type %sBO struct {%s
}

`, rlStructName, generateRLBOFields(rlTable)))

		// 生成RL表的ListBO结构定义（仅当有list字段时）
		if helper.ShouldIncludeRLInList(rlTable) {
			listBoDefinitionsBuf.WriteString(fmt.Sprintf(`type %sListBO struct {%s
}

`, rlStructName, generateRLListBOFields(rlTable)))
		}
	}

	// 生成BO字段定义
	for _, rlTable := range rlTables {
		rlStructName := helper.GetStructName(rlTable.Name)

		// 为详情BO添加RL表字段（总是包含）
		boFieldsBuf.WriteString(fmt.Sprintf("\t%ss []*%sBO `json:\"%ss,omitempty\"`\n",
			rlStructName, rlStructName, helper.GetVarName(rlTable.Name)))

		// 为列表BO添加RL表字段（仅当有list字段时，使用ListBO）
		if helper.ShouldIncludeRLInList(rlTable) {
			listBoFieldsBuf.WriteString(fmt.Sprintf("\t%ss []*%sListBO `json:\"%ss,omitempty\"`\n",
				rlStructName, rlStructName, helper.GetVarName(rlTable.Name)))
		}
	}

	// 生成Model到BO的转换逻辑
	for _, rlTable := range rlTables {
		rlStructName := helper.GetStructName(rlTable.Name)
		rlFieldName := rlStructName + "s"

		// 详情BO转换（总是包含）
		assignModelToBOBuf.WriteString(fmt.Sprintf(`
		%s: func() []*%sBO {
			result := make([]*%sBO, len(m.%s))
			for i, rl := range m.%s {
				result[i] = &%sBO{%s}
			}
			return result
		}(),`, rlFieldName, rlStructName, rlStructName, rlFieldName, rlFieldName, rlStructName,
			generateRLBOAssignment(rlTable)))

		// 列表BO转换（仅当有list字段时，使用ListBO）
		if helper.ShouldIncludeRLInList(rlTable) {
			assignModelToListBOBuf.WriteString(fmt.Sprintf(`
			%s: func() []*%sListBO {
				result := make([]*%sListBO, len(ms[i].%s))
				for j, rl := range ms[i].%s {
					result[j] = &%sListBO{%s}
				}
				return result
			}(),`, rlFieldName, rlStructName, rlStructName, rlFieldName, rlFieldName, rlStructName,
				generateRLListBOAssignment(rlTable)))
		}
	}

	// 生成RL表业务方法
	for _, rlTable := range rlTables {
		rlStructName := helper.GetStructName(rlTable.Name)
		mainTableStructName := helper.GetStructName(table.Name)

		// 生成Add方法
		methodsBuf.WriteString(fmt.Sprintf(`
func (b *BizService) Add%s(ctx context.Context, mainId uint64, data *%sBO) error {
	log := contexts.GetLogger(ctx).
		WithField("mainId", mainId).
		WithField("data", jgstr.JsonEncode(data))
	
	rl := &model.%s{%s
	}
	
	err := b.%sRepo.Add%s(ctx, mainId, rl)
	if err != nil {
		log.WithError(err).Error("fail to add %s")
		return cerror.Internal(err.Error())
	}
	return nil
}
`, rlStructName, rlStructName, rlStructName,
			generateBOToModelAssignment(rlTable), mainTableStructName, rlStructName, rlTable.Name))

		// 生成Remove方法
		methodsBuf.WriteString(fmt.Sprintf(`
func (b *BizService) Remove%s(ctx context.Context, mainId uint64, rlId uint64) error {
	log := contexts.GetLogger(ctx).
		WithField("mainId", mainId).
		WithField("rlId", rlId)
	
	err := b.%sRepo.Remove%s(ctx, mainId, rlId)
	if err != nil {
		log.WithError(err).Error("fail to remove %s")
		return cerror.Internal(err.Error())
	}
	return nil
}
`, rlStructName, mainTableStructName, rlStructName, rlTable.Name))

		// 生成Get方法
		methodsBuf.WriteString(fmt.Sprintf(`
func (b *BizService) GetAll%s(ctx context.Context, mainId uint64) ([]*%sBO, error) {
	log := contexts.GetLogger(ctx).
		WithField("mainId", mainId)
	
	rls, err := b.%sRepo.GetAll%s(ctx, mainId)
	if err != nil {
		log.WithError(err).Error("fail to get all %s")
		return nil, cerror.Internal(err.Error())
	}
	
	result := make([]*%sBO, len(rls))
	for i, rl := range rls {
		result[i] = &%sBO{%s}
	}
	return result, nil
}
`, rlStructName, rlStructName, mainTableStructName, rlStructName,
			rlTable.Name, rlStructName, rlStructName, generateRLBOAssignment(rlTable)))
	}

	// 生成事务中的RL表创建逻辑
	createInTransactionCode := generateRLCreateInTransaction(table)

	// 生成级联删除逻辑
	cascadeDeleteCode := generateRLCascadeDeleteInBiz(table)

	// 替换占位符
	*code = strings.ReplaceAll(*code, template.PH_RL_BO_DEFINITIONS, boDefinitionsBuf.String())
	*code = strings.ReplaceAll(*code, template.PH_RL_LIST_BO_DEFINITIONS, listBoDefinitionsBuf.String())
	*code = strings.ReplaceAll(*code, template.PH_RL_BO_FIELDS, boFieldsBuf.String())
	*code = strings.ReplaceAll(*code, template.PH_RL_LIST_BO_FIELDS, listBoFieldsBuf.String())
	*code = strings.ReplaceAll(*code, template.PH_RL_ASSIGN_MODEL_TO_BO, assignModelToBOBuf.String())
	*code = strings.ReplaceAll(*code, template.PH_RL_ASSIGN_MODEL_TO_LIST_BO, assignModelToListBOBuf.String())
	*code = strings.ReplaceAll(*code, template.PH_RL_CREATE_IN_TRANSACTION, createInTransactionCode)
	*code = strings.ReplaceAll(*code, template.PH_RL_CASCADE_DELETE_IN_BIZ, cascadeDeleteCode)
	*code = strings.ReplaceAll(*code, template.PH_RL_METHODS, methodsBuf.String())
}

func clearRLPlaceholders(code *string) {
	*code = strings.ReplaceAll(*code, template.PH_RL_BO_DEFINITIONS, "")
	*code = strings.ReplaceAll(*code, template.PH_RL_LIST_BO_DEFINITIONS, "")
	*code = strings.ReplaceAll(*code, template.PH_RL_BO_FIELDS, "")
	*code = strings.ReplaceAll(*code, template.PH_RL_LIST_BO_FIELDS, "")
	*code = strings.ReplaceAll(*code, template.PH_RL_ASSIGN_MODEL_TO_BO, "")
	*code = strings.ReplaceAll(*code, template.PH_RL_ASSIGN_MODEL_TO_LIST_BO, "")
	*code = strings.ReplaceAll(*code, template.PH_RL_CREATE_IN_TRANSACTION, "")
	*code = strings.ReplaceAll(*code, template.PH_RL_CASCADE_DELETE_IN_BIZ, "")
	*code = strings.ReplaceAll(*code, template.PH_RL_METHODS, "")
}

func generateRLBOFields(table *model.Table) string {
	var buf strings.Builder
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}

		gotype, err := helper.GetGoType(col)
		if err != nil {
			continue
		}

		// 处理必填性：如果字段不是必填的且Go类型不是可空的，则添加指针
		if !col.IsRequired && !helper.IsGoTypeNullable(gotype) {
			gotype = "*" + gotype
		}

		jsonTag := helper.GetVarName(col.Name)
		comment := col.Comment
		buf.WriteString(fmt.Sprintf("\n\t%s %s `json:\"%s\"` // %s",
			helper.GetTableColName(col.Name), gotype, jsonTag, comment))
	}
	return buf.String()
}

func generateRLBOAssignment(table *model.Table) string {
	var buf strings.Builder
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}
		buf.WriteString(fmt.Sprintf("\n\t\t\t%s: rl.%s,",
			helper.GetTableColName(col.Name), helper.GetTableColName(col.Name)))
	}
	return buf.String()
}

func generateBOToModelAssignment(table *model.Table) string {
	var buf strings.Builder
	for _, col := range table.Columns {
		if col.IsPrimaryKey {
			continue // 主键通常由数据库自动生成
		}
		if col.IsHidden {
			continue
		}
		if !col.IsAlterable {
			continue // 只考虑alter=true的字段
		}

		// 跳过主外键字段，这些会由repo层自动处理
		isMainForeignKey := false
		for _, fk := range col.ForeignKeys {
			if fk.IsMain {
				isMainForeignKey = true
				break
			}
		}
		if !isMainForeignKey {
			buf.WriteString(fmt.Sprintf("\n\t\t%s: data.%s,",
				helper.GetTableColName(col.Name), helper.GetTableColName(col.Name)))
		}
	}
	return buf.String()
}

func generateRLListBOFields(table *model.Table) string {
	var buf strings.Builder
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}
		if !col.IsList {
			continue // 只包含list=true的字段
		}

		gotype, err := helper.GetGoType(col)
		if err != nil {
			continue
		}

		// 处理必填性：如果字段不是必填的且Go类型不是可空的，则添加指针
		if !col.IsRequired && !helper.IsGoTypeNullable(gotype) {
			gotype = "*" + gotype
		}

		jsonTag := helper.GetVarName(col.Name)
		comment := col.Comment
		buf.WriteString(fmt.Sprintf("\n\t%s %s `json:\"%s\"` // %s",
			helper.GetTableColName(col.Name), gotype, jsonTag, comment))
	}
	return buf.String()
}

func generateRLListBOAssignment(table *model.Table) string {
	var buf strings.Builder
	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}
		if !col.IsList {
			continue // 只包含list=true的字段
		}
		buf.WriteString(fmt.Sprintf("\n\t\t\t%s: rl.%s,",
			helper.GetTableColName(col.Name), helper.GetTableColName(col.Name)))
	}
	return buf.String()
}

func generateRLCreateInTransaction(table *model.Table) string {
	var buf strings.Builder
	rlTables := helper.GetMainTableRLs(table, table.Database.Tables)

	for _, rlTable := range rlTables {
		rlStructName := helper.GetStructName(rlTable.Name)
		rlFieldName := rlStructName + "s"

		buf.WriteString(fmt.Sprintf(`
		// 创建%s数据
		for _, data := range obj.%s {
			rlModel := &model.%s{%s
			}
			err = b.%sRepo.Add%s(txCtx, m.ID, rlModel)
			if err != nil {
				log.WithError(err).Error("fail to add %s")
				return cerror.Internal(err.Error())
			}
			// 将新创建的RL记录添加到主表model中，确保返回的BO包含完整数据
			m.%s = append(m.%s, rlModel)
		}`, rlTable.Comment, rlFieldName, rlStructName,
			generateBOToModelAssignment(rlTable), helper.GetStructName(table.Name), rlStructName,
			rlTable.Name, rlFieldName, rlFieldName))
	}

	return buf.String()
}

func generateRLCascadeDeleteInBiz(table *model.Table) string {
	var buf strings.Builder
	rlTables := helper.GetMainTableRLs(table, table.Database.Tables)
	mainTableStructName := helper.GetStructName(table.Name)

	for _, rlTable := range rlTables {
		rlStructName := helper.GetStructName(rlTable.Name)

		// 检查是否有设置auto_remove的外键
		hasAutoRemove := false
		for _, column := range rlTable.Columns {
			for _, fk := range column.ForeignKeys {
				if fk.Table == table.Name && fk.AutoRemove {
					hasAutoRemove = true
					break
				}
			}
			if hasAutoRemove {
				break
			}
		}

		// 只处理设置了auto_remove的RL表
		if !hasAutoRemove {
			continue
		}

		buf.WriteString(fmt.Sprintf(`
		// 级联删除%s
		err = b.%sRepo.RemoveAll%s(txCtx, id)
		if err != nil {
			log.WithError(err).Error("fail to remove %ss")
			return cerror.Internal(err.Error())
		}`,
			rlTable.Comment,
			mainTableStructName, rlStructName,
			helper.GetVarName(rlTable.Name)))
	}

	return buf.String()
}

func genBRBizMethods(code *string, table *model.Table) {
	// 构建表名到表的映射
	tableNameToTable := make(map[string]*model.Table)
	for _, t := range table.Database.Tables {
		tableNameToTable[t.Name] = t
	}

	// 获取当前表的所有BR表关系
	brTables := helper.GetMainTableBRs(table, table.Database.Tables)

	if len(brTables) == 0 {
		clearBRPlaceholders(code)
		return
	}

	var methodsBuf strings.Builder

	// 为每个BR表生成关系管理方法
	for _, brTable := range brTables {
		// 识别BR表连接的两个数据表
		brRelated := helper.IdentifyBRRelatedTables(brTable, tableNameToTable)
		if brRelated == nil {
			continue
		}

		// 获取对方表
		otherTable := helper.GetBROtherTable(brTable, table, tableNameToTable)
		if otherTable == nil {
			continue
		}

		otherTableStructName := helper.GetStructName(otherTable.Name)
		thisTableStructName := helper.GetStructName(table.Name)

		// 生成Get{CurrentTable}Related{OtherTable}List方法，直接复用目标表的ListOption
		methodsBuf.WriteString(fmt.Sprintf(`
func (b *BizService) Get%sRelated%sList(ctx context.Context, %sId uint64, opt *%sListOption) ([]*%sListBO, int64, error) {
	log := contexts.GetLogger(ctx).
		WithField("%sId", %sId).
		WithField("opt", jgstr.JsonEncode(opt))
	
	// 转换biz层option为repo层option
	related%ss, total, err := b.%sRepo.GetRelated%sList(ctx, %sId, &option.Related%sListOption{
		Pagination: opt.Pagination,
		Order: opt.Order,%s
		Select: []interface{}{%s
		},
	})
	if err != nil {
		log.WithError(err).Error("fail to get related %s list")
		return nil, 0, cerror.Internal(err.Error())
	}
	
	// 转换为ListBO
	result, err := b.To%sListBO(ctx, related%ss)
	if err != nil {
		log.WithError(err).Error("fail to convert %s models to %sListBO")
		return nil, 0, cerror.Internal(err.Error())
	}
	return result, total, nil
}
`, thisTableStructName, otherTableStructName, // Get{CurrentTable}Related{OtherTable}List
			helper.GetVarName(table.Name), otherTableStructName, otherTableStructName, // 参数
			helper.GetVarName(table.Name), helper.GetVarName(table.Name), // log字段
			otherTableStructName, thisTableStructName, otherTableStructName, helper.GetVarName(table.Name), otherTableStructName, // repo调用
			generateBRBizFilterAssignmentDirect(otherTable),           // filter转换
			generateBRBizSelectList(otherTable, otherTableStructName), // select字段（不带表名前缀）
			otherTable.Name,                            // 错误日志
			otherTableStructName, otherTableStructName, // BO转换
			otherTable.Name, otherTableStructName)) // BO转换错误日志
	}

	// 替换占位符
	*code = strings.ReplaceAll(*code, template.PH_BR_OPTIONS, "")
	*code = strings.ReplaceAll(*code, template.PH_BR_METHODS, methodsBuf.String())
}

func clearBRPlaceholders(code *string) {
	*code = strings.ReplaceAll(*code, template.PH_BR_OPTIONS, "")
	*code = strings.ReplaceAll(*code, template.PH_BR_METHODS, "")
}

// generateBRBizFilterAssignmentDirect 生成biz层filter到repo层filter的转换赋值，直接使用目标表的ListOption
func generateBRBizFilterAssignmentDirect(otherTable *model.Table) string {
	var buf strings.Builder
	hasFilterCol := false
	for _, col := range otherTable.Columns {
		if col.IsFilter {
			hasFilterCol = true
			break
		}
	}
	if !hasFilterCol {
		return ""
	}

	buf.WriteString(fmt.Sprintf("\n\t\tFilter: &option.Related%sFilterOption{", helper.GetStructName(otherTable.Name)))
	for _, col := range otherTable.Columns {
		if col.IsFilter {
			buf.WriteString(fmt.Sprintf("\n\t\t\t%s: opt.Filter.%s,",
				helper.GetTableColName(col.Name),
				helper.GetTableColName(col.Name),
			))
		}
	}
	buf.WriteString("\n\t\t},")
	return buf.String()
}

// generateBRBizSelectList 生成target table的select字段列表（不带表名前缀，供biz层使用）
func generateBRBizSelectList(table *model.Table, otherTableStructName string) string {
	var buf strings.Builder
	tableStructName := helper.GetStructName(table.Name)

	for _, col := range table.Columns {
		if col.IsHidden {
			continue
		}
		if !col.IsList {
			continue
		}
		// 生成不带表名前缀的select字段：model.ColRoleName（repo层会自动补全表名前缀）
		buf.WriteString(fmt.Sprintf("\n\t\t\tmodel.Col%s%s,",
			tableStructName,
			helper.GetTableColName(col.Name),
		))
	}
	return buf.String()
}
