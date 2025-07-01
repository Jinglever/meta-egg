package modelgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"meta-egg/internal/domain/helper"
	template "meta-egg/internal/domain/model_generator/template"
	"meta-egg/internal/model"
	"meta-egg/internal/repo"

	jgstr "github.com/Jinglever/go-string"
	log "github.com/sirupsen/logrus"
)

// 生成model文件
// relativeDir2NeedConfirm: key是相对于工程根目录的路径，值是布尔值，代表是否需要确认；注意，不会递归查看子目录
func Generate(codeDir string, project *model.Project, dbOper repo.DBOperator) (relativeDir2NeedConfirm map[string]bool, err error) {
	relativeDir2NeedConfirm = map[string]bool{
		filepath.Join("gen", "model"): false,
	}
	// 创建目录
	for dir := range relativeDir2NeedConfirm {
		path := filepath.Join(codeDir, dir)
		if err = os.MkdirAll(path, 0755); err != nil {
			log.Errorf("failed to mkdir %s: %v", dir, err)
			return
		}
	}

	if project.Database != nil {
		modelDir := filepath.Join(codeDir, "gen", "model")
		for _, table := range project.Database.Tables {
			err = generateForTable(filepath.Join(modelDir, table.Name+".go"), table, dbOper)
			if err != nil {
				log.Errorf("generate for table failed: %v", err)
				return
			}
		}
	}
	return
}

func generateForTable(path string, table *model.Table, dbOper repo.DBOperator) error {
	f, err := os.Create(path)
	if err != nil {
		log.Errorf("create file failed: %v", err)
		return err
	}
	defer f.Close()

	// template
	code := template.TplTable
	code = helper.AddHeaderNoEdit(code)

	var buf strings.Builder
	var imports map[string]bool = make(map[string]bool)

	tableName := helper.GetTableColName(table.Name)

	// const column name
	for _, col := range table.Columns {
		if col.IsPrimaryKey && col.Comment == "" {
			buf.WriteString(fmt.Sprintf(`
			Col%s%s = "%s" //nolint`, tableName, helper.GetTableColName(col.Name), col.Name))
		} else {
			buf.WriteString(fmt.Sprintf(`
			Col%s%s = "%s" // %s`, tableName, helper.GetTableColName(col.Name), col.Name, col.Comment))
		}
	}
	code = strings.ReplaceAll(code, template.PH_CONST_COL_LIST, buf.String())
	buf.Reset()

	// struct
	for _, column := range table.Columns {
		buf.WriteString("\n")
		// soft delete
		if column.ExtType == model.ColumnExtType_TIME_DELETE {
			typ := "gorm.DeletedAt"
			imports["gorm.io/gorm"] = true
			buf.WriteString(fmt.Sprintf("	 %s %s `gorm:\"column:%s\"` // %s", helper.GetTableColName(column.Name), typ, column.Name, column.Comment))
			continue
		} else if column.ExtType == model.ColumnExtType_TIME_DELETE2 {
			typ := "soft_delete.DeletedAt"
			imports["gorm.io/plugin/soft_delete"] = true
			buf.WriteString(fmt.Sprintf("	 %s %s `gorm:\"column:%s\"` // %s", helper.GetTableColName(column.Name), typ, column.Name, column.Comment))
			continue
		}

		typ, err := helper.GetGoType(column)
		if err != nil {
			log.Errorf("get go type failed: %v", err)
			return err
		}
		if column.IsPrimaryKey && column.Comment == "" {
			buf.WriteString(fmt.Sprintf("	 %s %s `gorm:\"column:%s\"`", helper.GetTableColName(column.Name), typ, column.Name))
		} else {
			if column.IsRequired {
				buf.WriteString(fmt.Sprintf("	 %s %s `gorm:\"column:%s\"` // %s", helper.GetTableColName(column.Name), typ, column.Name, column.Comment))
			} else {
				if !helper.IsGoTypeNullable(typ) {
					buf.WriteString(fmt.Sprintf("	 %s *%s `gorm:\"column:%s\"` // %s [default NULL]", helper.GetTableColName(column.Name), typ, column.Name, column.Comment))
				} else {
					buf.WriteString(fmt.Sprintf("	 %s %s `gorm:\"column:%s\"` // %s [default NULL]", helper.GetTableColName(column.Name), typ, column.Name, column.Comment))
				}
			}
		}
		if typ == "time.Time" {
			imports["time"] = true
		}
	}

	// Add RL associations for DATA tables
	if table.Type == model.TableType_DATA {
		rlTables := helper.GetMainTableRLs(table, table.Database.Tables)
		for _, rlTable := range rlTables {
			rlTableStructName := helper.GetTableColName(rlTable.Name)
			rlFieldName := helper.GetTableColName(rlTable.Name) + "s"
			buf.WriteString(fmt.Sprintf("\n	 %s []*%s `gorm:\"foreignKey:%s\"` // %s",
				rlFieldName, rlTableStructName, findForeignKeyColumnToMainTable(rlTable, table), rlTable.Comment))
		}
	}

	code = strings.ReplaceAll(code, template.PH_STRUCT_COL_LIST, buf.String())
	buf.Reset()

	// AfterFind
	needAfterFind := false
	for _, col := range table.Columns {
		if table.Database.Type == model.DBType_PG {
			if col.Type == model.ColumnType_TIME ||
				col.Type == model.ColumnType_DATETIME ||
				col.Type == model.ColumnType_DATE {
				needAfterFind = true
				break
			}
		}
	}
	if needAfterFind {
		code = strings.ReplaceAll(code, template.PH_TPL_AFTER_FIND, template.TplAfterFind)
		genCorrectTimezone(&code, table)
	} else {
		code = strings.ReplaceAll(code, template.PH_TPL_AFTER_FIND, "")
	}

	// const meta ids
	eCols := helper.GetExtTypeCols(table)
	if table.Type != model.TableType_META {
		code = strings.ReplaceAll(code, template.PH_TPL_CONST_META_IDS, "")
	} else {
		metaRecords, err := helper.GetMetaRecords(table, eCols, dbOper)
		if err != nil {
			log.Errorf("get meta records failed: %v", err)
			return err
		}
		genConstMetaIDs(&code, table, eCols, metaRecords)
	}

	// import
	for k := range imports {
		buf.WriteString(fmt.Sprintf("\n	\"%s\"", k))
	}
	code = strings.ReplaceAll(code, template.PH_IMPORTS, buf.String())
	buf.Reset()

	code = strings.ReplaceAll(code, template.PH_DB_TABLE_NAME, table.Name)
	code = strings.ReplaceAll(code, template.PH_STRUCT_TABLE_NAME, tableName)
	code = strings.ReplaceAll(code, template.PH_TABLE_COMMENT, table.Comment)
	code = strings.ReplaceAll(code, template.PH_GO_MODULE, table.Database.Project.GoModule)

	// go format
	formatted, err := jgstr.FormatGo([]byte(code))
	if err != nil {
		log.Errorf("format source failed: %v\n%s", err, code)
		return err
	}
	f.Write(formatted)
	return nil
}

func genConstMetaIDs(code *string, table *model.Table, eCols *helper.ExtTypeCols, metaRecords []map[string]interface{}) {
	if len(metaRecords) == 0 {
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONST_META_IDS, "")
	} else {
		var buf strings.Builder
		// const meta ids
		for _, record := range metaRecords {
			if eCols.SemanticDesc == nil {
				buf.WriteString(fmt.Sprintf("\n	Meta%s%s uint64 = %d",
					helper.GetStructName(table.Name),
					helper.GetTableColName(record[eCols.Semantic.Name].(string)),
					record[table.PrimaryColumn.Name]))
			} else {
				buf.WriteString(fmt.Sprintf("\n	Meta%s%s uint64 = %d // %s",
					helper.GetStructName(table.Name),
					helper.GetTableColName(record[eCols.Semantic.Name].(string)),
					record[table.PrimaryColumn.Name],
					record[eCols.SemanticDesc.Name]))
			}
		}
		*code = strings.ReplaceAll(*code, template.PH_TPL_CONST_META_IDS, template.TplConstMetaIDs)
		*code = strings.ReplaceAll(*code, template.PH_CONST_META_ID_LIST, buf.String())
	}
}

/*
e.g.
t.CreatedAt = gormx.CorrectTimezone(t.CreatedAt)

	if t.UpdatedAt != nil {
		tmp := gormx.CorrectTimezone(*t.UpdatedAt)
		t.UpdatedAt = &tmp
	}

t.DeletedAt.Time = gormx.CorrectTimezone(t.DeletedAt.Time)
*/
func genCorrectTimezone(code *string, table *model.Table) {
	var buf strings.Builder
	for _, col := range table.Columns {
		if col.ExtType == model.ColumnExtType_TIME_DELETE {
			buf.WriteString(fmt.Sprintf("\n	t.%s.Time = gormx.CorrectTimezone(t.%s.Time)", helper.GetTableColName(col.Name), helper.GetTableColName(col.Name)))
			continue
		}
		if col.Type == model.ColumnType_TIME ||
			col.Type == model.ColumnType_DATETIME ||
			col.Type == model.ColumnType_DATE {
			if col.IsRequired {
				buf.WriteString(fmt.Sprintf("\n	t.%s = gormx.CorrectTimezone(t.%s)", helper.GetTableColName(col.Name), helper.GetTableColName(col.Name)))
			} else {
				buf.WriteString(fmt.Sprintf("\n	if t.%s != nil {", helper.GetTableColName(col.Name)))
				buf.WriteString(fmt.Sprintf("\n		tmp := gormx.CorrectTimezone(*t.%s)", helper.GetTableColName(col.Name)))
				buf.WriteString(fmt.Sprintf("\n		t.%s = &tmp", helper.GetTableColName(col.Name)))
				buf.WriteString("\n	}")
			}
		}
	}
	*code = strings.ReplaceAll(*code, template.PH_CORRECT_TIMEZONE, buf.String())
}

// findForeignKeyColumnToMainTable 找到RL表中指向主表的主外键字段名
func findForeignKeyColumnToMainTable(rlTable *model.Table, mainTable *model.Table) string {
	// 首先查找标记为主外键的字段
	for _, column := range rlTable.Columns {
		for _, foreignKey := range column.ForeignKeys {
			if foreignKey.Table == mainTable.Name && foreignKey.IsMain {
				return helper.GetTableColName(column.Name)
			}
		}
	}

	// 如果没有明确标记的主外键，查找第一个指向主表的外键（向后兼容）
	for _, column := range rlTable.Columns {
		for _, foreignKey := range column.ForeignKeys {
			if foreignKey.Table == mainTable.Name {
				return helper.GetTableColName(column.Name)
			}
		}
	}

	return ""
}
