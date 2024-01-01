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
	} else if table.Type == model.TableType_META {
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_GET_LIST, template.TplFuncGetList)
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_UPDATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_DELETE, "")
	} else {
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_CREATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_GET_LIST, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_UPDATE, "")
		*code = strings.ReplaceAll(*code, template.PH_TPL_FUNC_DELETE, "")
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
