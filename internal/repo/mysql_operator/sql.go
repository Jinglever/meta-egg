package mysqloperator

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"meta-egg/internal/model"

	log "github.com/sirupsen/logrus"
)

/*
 * 输出更新Schema的SQL
 * @param targetDBSchema 目标数据库的schema
 * @param createSQLWriter 输出创建数据库的SQL
 * @param incSQLWriter 输出更新数据库的增量SQL
 * @param metaDataSQLWriter 输出数据库的meta表的数据的INSERT SQL
 */
func (t *MySQLOperator) OutputSQLForSchemaUpdating(targetDBSchema *model.Database,
	createSQLWriter io.StringWriter,
	incSQLWriter io.StringWriter,
	metaDataSQLWriter io.StringWriter,
) error {
	var wg sync.WaitGroup

	// create SQL
	wg.Add(1)
	go func() {
		defer wg.Done()
		createSQLWriter.WriteString("SET FOREIGN_KEY_CHECKS=0;\n")
		createSQLWriter.WriteString(fmt.Sprintf("CREATE DATABASE `%s` DEFAULT CHARACTER SET %s COLLATE %s;\n",
			targetDBSchema.Name, targetDBSchema.Charset, targetDBSchema.Collate))
		createSQLWriter.WriteString("USE `" + targetDBSchema.Name + "`;\n")
		for _, table := range targetDBSchema.Tables {
			t.outputCreateTableSQL(table, createSQLWriter)
		}
	}()

	// inc SQL
	wg.Add(1)
	go func() {
		defer wg.Done()
		curDBSchema, err := t.GetCurDBSchema()
		if err != nil {
			log.Errorf("get current database schema failed: %v", err)
			return
		}

		incSQLWriter.WriteString("SET FOREIGN_KEY_CHECKS=0;\n")

		// database的以下属性只能手工变更：name、charset、collate
		// if curDBSchema.Name != targetDBSchema.Name {
		// 	log.Warnf("need modify manually: database.name")
		// }
		if curDBSchema.Charset != targetDBSchema.Charset {
			log.Warnf("need modify manually: database.charset")
		}
		if curDBSchema.Collate != targetDBSchema.Collate {
			log.Warnf("need modify manually: database.collate")
		}

		// tables
		// todo table.name modify dispend on dtd rule support
		// todo column.name modify dispend on dtd rule support
		for _, table := range curDBSchema.Tables {
			if _, ok := targetDBSchema.TableNameToTable[table.Name]; !ok {
				// rm tables
				t.outputRemoveTableSQL(table, incSQLWriter)
			}
		}
		for _, table := range targetDBSchema.Tables {
			if _, ok := curDBSchema.TableNameToTable[table.Name]; !ok {
				// add tables
				t.outputCreateTableSQL(table, incSQLWriter)
				continue
			} else {
				// modify
				t.outputModifyTableSQL(curDBSchema.TableNameToTable[table.Name], table, incSQLWriter)
			}
		}
	}()

	// meta data sql
	wg.Add(1)
	go func() {
		defer wg.Done()
		metaDataSQLWriter.WriteString("SET FOREIGN_KEY_CHECKS=0;\n")
		metaDataSQLWriter.WriteString("USE `" + targetDBSchema.Name + "`;\n")
		for _, table := range targetDBSchema.Tables {
			if table.Type != model.TableType_META {
				continue
			}

			// get meta data
			allRecords, err := t.GetAllRecords(table.Name, table.PrimaryColumn.Name, []string{})
			if err != nil {
				log.Warnf("get meta data failed: %v", err)
				continue
			}

			if len(allRecords) == 0 {
				continue
			}

			// fetch all colNames into []string from allRecords[0]
			var colNames []string
			for key := range allRecords[0] {
				colNames = append(colNames, key)
			}
			// sort colNames using standard package
			sort.Strings(colNames)

			// write table name as comment
			metaDataSQLWriter.WriteString(fmt.Sprintf("-- %s\n", table.Name))
			// write batch insert sql, specify column name
			metaDataSQLWriter.WriteString(fmt.Sprintf("INSERT INTO `%s` (`%s`) VALUES\n",
				table.Name, strings.Join(colNames, "`,`")))
			for i, record := range allRecords {
				metaDataSQLWriter.WriteString("(")
				for j, colName := range colNames {
					if record[colName] == nil {
						metaDataSQLWriter.WriteString("NULL")
					} else {
						if v, ok := record[colName].(string); ok {
							metaDataSQLWriter.WriteString(fmt.Sprintf("'%v'", strings.ReplaceAll(v, "'", "\\'")))
						} else {
							metaDataSQLWriter.WriteString(fmt.Sprintf("'%v'", record[colName]))
						}
					}
					if j != len(colNames)-1 {
						metaDataSQLWriter.WriteString(", ")
					}
				}
				metaDataSQLWriter.WriteString(")")
				if i != len(allRecords)-1 {
					metaDataSQLWriter.WriteString(",\n")
				}
			}
			metaDataSQLWriter.WriteString(";\n\n")
		}
	}()

	wg.Wait()
	return nil
}

func (t *MySQLOperator) outputCreateTableSQL(table *model.Table, writer io.StringWriter) {
	// create table
	writer.WriteString(fmt.Sprintf("CREATE TABLE `%s` (\n", table.Name))

	// primary key
	writer.WriteString(fmt.Sprintf("  %s", t.outputColumnSQL(table.PrimaryColumn)))

	// normal column
	for _, column := range table.Columns {
		if column.Name == table.PrimaryColumn.Name {
			continue
		}

		writer.WriteString(",\n")
		writer.WriteString(fmt.Sprintf("  %s", t.outputColumnSQL(column)))
	}

	// single index
	for _, column := range table.Columns {
		if column.IsFullText && len(column.ForeignKeys) == 0 {
			writer.WriteString(",\n")
			writer.WriteString(fmt.Sprintf("  %s", t.outputSingleFullTextSQL(column.Name, column.IndexAlias)))
		}
		if column.IsUnique && len(column.ForeignKeys) == 0 {
			writer.WriteString(",\n")
			writer.WriteString(fmt.Sprintf("  %s", t.outputSingleUniqueSQL(column.Name, column.IndexAlias)))
		}
		if column.IsIndex && len(column.ForeignKeys) == 0 {
			writer.WriteString(",\n")
			writer.WriteString(fmt.Sprintf("  %s", t.outputSingleIndexSQL(column.Name, column.IndexAlias)))
		}
	}

	// multiple fulltext index
	for _, fulltext := range table.FullText {
		writer.WriteString(",\n")
		writer.WriteString(fmt.Sprintf("  %s", t.outputFullTextSQL(fulltext)))
	}

	// multiple unique index
	for _, unique := range table.Unique {
		writer.WriteString(",\n")
		writer.WriteString(fmt.Sprintf("  %s", t.outputUniqueSQL(unique)))
	}

	// multiple index
	for _, index := range table.Indexes {
		writer.WriteString(",\n")
		writer.WriteString(fmt.Sprintf("  %s", t.outputIndexSQL(index)))
	}

	// foreign key
	for _, column := range table.Columns {
		if len(column.ForeignKeys) > 0 {
			writer.WriteString(",\n")
			writer.WriteString(fmt.Sprintf("  %s", t.outputForeignKeySQL(column.ForeignKeys[0])))
		}
	}

	// table comment
	writer.WriteString("\n")
	writer.WriteString(fmt.Sprintf(") ENGINE=INNODB DEFAULT CHARSET=%s COLLATE=%s COMMENT='%s';\n",
		table.Charset, table.Collate, table.Comment))
}

func (t *MySQLOperator) outputRemoveTableSQL(table *model.Table, writer io.StringWriter) {
	if !t.IgnoreTables[table.Name] {
		writer.WriteString(fmt.Sprintf("DROP TABLE `%s`;\n", table.Name))
	}
}

func (t *MySQLOperator) outputModifyTableSQL(curTable *model.Table, targetTable *model.Table, writer io.StringWriter) {
	// table attributes
	attrs := make([]string, 0)
	if curTable.Charset != targetTable.Charset {
		attrs = append(attrs, fmt.Sprintf("  CHARSET = %s", targetTable.Charset))
	}
	if curTable.Collate != targetTable.Collate {
		attrs = append(attrs, fmt.Sprintf("  COLLATE = %s", targetTable.Collate))
	}
	if curTable.Comment != targetTable.Comment {
		attrs = append(attrs, fmt.Sprintf("  COMMENT = '%s'", targetTable.Comment))
	}
	if len(attrs) > 0 {
		writer.WriteString(fmt.Sprintf("ALTER TABLE `%s`\n", targetTable.Name))
		for i := 0; i < len(attrs)-1; i++ {
			writer.WriteString(attrs[i] + ",\n")
		}
		writer.WriteString(attrs[len(attrs)-1] + ";\n")
	}
	// multi-index
	for key, idx := range curTable.ColNamesToIndex {
		if _, ok := targetTable.ColNamesToIndex[key]; !ok {
			// rm index
			if idx.Alias != "" {
				writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` DROP INDEX %s;\n", targetTable.Name, idx.Alias))
			}
		}
	}
	for key, idx := range curTable.ColNamesToUnique {
		if _, ok := targetTable.ColNamesToUnique[key]; !ok {
			// rm unique
			if idx.Alias != "" {
				writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` DROP INDEX %s;\n", targetTable.Name, idx.Alias))
			}
		}
	}
	for key, idx := range curTable.ColNamesToFullText {
		if _, ok := targetTable.ColNamesToFullText[key]; !ok {
			// rm fulltext
			if idx.Alias != "" {
				writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` DROP INDEX %s;\n", targetTable.Name, idx.Alias))
			}
		}
	}
	// column
	for _, column := range curTable.ColNameToColumn {
		if _, ok := targetTable.ColNameToColumn[column.Name]; !ok {
			// rm index & foreign key first if exists
			// drop foreign key first
			if len(column.ForeignKeys) > 0 && column.ForeignKeys[0].Alias != "" {
				writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` DROP FOREIGN KEY %s;\n", targetTable.Name, column.ForeignKeys[0].Alias))
			}
			if column.IsIndex && column.IndexAlias != "" {
				writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` DROP INDEX %s;\n", targetTable.Name, column.IndexAlias))
			}
			if column.IsUnique && column.IndexAlias != "" {
				writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` DROP INDEX %s;\n", targetTable.Name, column.IndexAlias))
			}
			if column.IsFullText && column.IndexAlias != "" {
				writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` DROP INDEX %s;\n", targetTable.Name, column.IndexAlias))
			}
			// rm column
			writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`;\n", targetTable.Name, column.Name))
		}
	}
	for _, column := range targetTable.ColNameToColumn {
		curCol, ok := curTable.ColNameToColumn[column.Name]
		if !ok {
			// add column
			writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` ADD %s;\n", targetTable.Name, t.outputColumnSQL(column)))
			if column.IsFullText && len(column.ForeignKeys) == 0 {
				writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` ADD %s;\n", targetTable.Name, t.outputSingleFullTextSQL(column.Name, column.IndexAlias)))
			}
			if column.IsUnique && len(column.ForeignKeys) == 0 {
				writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` ADD %s;\n", targetTable.Name, t.outputSingleUniqueSQL(column.Name, column.IndexAlias)))
			}
			if column.IsIndex && len(column.ForeignKeys) == 0 {
				writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` ADD %s;\n", targetTable.Name, t.outputSingleIndexSQL(column.Name, column.IndexAlias)))
			}
			if len(column.ForeignKeys) > 0 {
				writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` ADD %s;\n", targetTable.Name, t.outputForeignKeySQL(column.ForeignKeys[0])))
			}
		} else {
			if column.Type != curCol.Type ||
				(column.Length != curCol.Length &&
					(model.StringColumnTypes[column.Type] ||
						column.Type == model.ColumnType_DECIMAL ||
						model.BinaryColumnTypes[column.Type])) ||
				column.Decimal != curCol.Decimal ||
				column.IsUnsigned != curCol.IsUnsigned ||
				column.IsRequired != curCol.IsRequired ||
				column.IsPrimaryKey != curCol.IsPrimaryKey ||
				column.Comment != curCol.Comment ||
				(column.DefaultEmpty != curCol.DefaultEmpty &&
					model.DefaultEmptyColTypes[column.Type]) ||
				(column.InsertDefault != curCol.InsertDefault &&
					!model.NoDefaultColTypes[column.Type] &&
					!(model.DefaultEmptyColTypes[column.Type] &&
						column.DefaultEmpty)) {
				// modify column
				writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` MODIFY %s;\n", targetTable.Name, t.outputColumnSQL(column)))
			}
			// single-index
			if curCol.IsIndex != column.IsIndex {
				if column.IsIndex && len(column.ForeignKeys) == 0 {
					writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` ADD %s;\n", targetTable.Name, t.outputSingleIndexSQL(column.Name, column.IndexAlias)))
				} else if curCol.IndexAlias != "" {
					writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`;\n", targetTable.Name, curCol.IndexAlias))
				}
			}
			if curCol.IsUnique != column.IsUnique {
				if column.IsUnique {
					writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` ADD %s;\n", targetTable.Name, t.outputSingleUniqueSQL(column.Name, column.IndexAlias)))
				} else if curCol.IndexAlias != "" {
					writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`;\n", targetTable.Name, curCol.IndexAlias))
				}
			}
			if curCol.IsFullText != column.IsFullText {
				if column.IsFullText {
					writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` ADD %s;\n", targetTable.Name, t.outputSingleFullTextSQL(column.Name, column.IndexAlias)))
				} else if curCol.IndexAlias != "" {
					writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`;\n", targetTable.Name, curCol.IndexAlias))
				}
			}
			// foreign key
			if len(curCol.ForeignKeys) != len(column.ForeignKeys) {
				if len(column.ForeignKeys) > 0 {
					writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` ADD %s;\n", targetTable.Name, t.outputForeignKeySQL(column.ForeignKeys[0])))
				} else if len(curCol.ForeignKeys) > 0 && curCol.ForeignKeys[0].Alias != "" {
					writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` DROP FOREIGN KEY %s;\n", targetTable.Name, curCol.ForeignKeys[0].Alias))
				}
			}
		}
	}
	// add multi-index
	for key, idx := range targetTable.ColNamesToIndex {
		if _, ok := curTable.ColNamesToIndex[key]; !ok {
			// add index
			writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` ADD %s;\n", targetTable.Name, t.outputIndexSQL(idx)))
		}
	}
	for key, idx := range targetTable.ColNamesToUnique {
		if _, ok := curTable.ColNamesToUnique[key]; !ok {
			// add unique
			writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` ADD %s;\n", targetTable.Name, t.outputUniqueSQL(idx)))
		}
	}
	for key, idx := range targetTable.ColNamesToFullText {
		if _, ok := curTable.ColNamesToFullText[key]; !ok {
			// add fulltext
			writer.WriteString(fmt.Sprintf("ALTER TABLE `%s` ADD %s;\n", targetTable.Name, t.outputFullTextSQL(idx)))
		}
	}
}

func (t *MySQLOperator) outputColumnSQL(column *model.Column) string {
	var sql string

	if column == column.Table.PrimaryColumn {
		// primary key
		sql += fmt.Sprintf("`%s`", column.Name)
		if column.Length > 0 {
			sql += fmt.Sprintf(" %s(%d)", column.Type, column.Length)
		} else {
			sql += fmt.Sprintf(" %s", column.Type)
		}
		if column.IsUnsigned {
			sql += " UNSIGNED"
		}
		sql += " AUTO_INCREMENT"
		sql += " PRIMARY KEY"

	} else {
		// normal column
		sql += fmt.Sprintf("`%s`", column.Name)
		if column.Type == model.ColumnType_DECIMAL {
			sql += fmt.Sprintf(" %s(%d,%d)", column.Type, column.Length, column.Decimal)
		} else {
			if column.Length > 0 {
				sql += fmt.Sprintf(" %s(%d)", column.Type, column.Length)
			} else {
				sql += fmt.Sprintf(" %s", column.Type)
			}
		}
		if column.IsUnsigned {
			sql += " UNSIGNED"
		}
		if column.IsRequired {
			sql += " NOT NULL"
		}
		if !model.NoDefaultColTypes[column.Type] {
			if column.DefaultEmpty && model.DefaultEmptyColTypes[column.Type] {
				sql += " DEFAULT ''"
			} else if column.InsertDefault != "" {
				sql += fmt.Sprintf(" DEFAULT '%s'", column.InsertDefault)
			}
		}
		if column.Comment != "" {
			sql += fmt.Sprintf(" COMMENT '%s'", column.Comment)
		}
	}

	return sql
}

func (t *MySQLOperator) outputIndexSQL(index *model.Index) string {
	var sql string
	items := make([]string, 0, len(index.IndexColumns))
	for _, column := range index.IndexColumns {
		items = append(items, fmt.Sprintf("`%s`", column.Name))
	}
	if index.Alias != "" {
		sql += fmt.Sprintf("INDEX %s (%s)", index.Alias, strings.Join(items, ", "))
	} else {
		sql += fmt.Sprintf("INDEX (%s)", strings.Join(items, ", "))
	}
	return sql
}

func (t *MySQLOperator) outputUniqueSQL(unique *model.Unique) string {
	var sql string
	items := make([]string, 0, len(unique.IndexColumns))
	for _, column := range unique.IndexColumns {
		items = append(items, fmt.Sprintf("`%s`", column.Name))
	}
	if unique.Alias != "" {
		sql += fmt.Sprintf("UNIQUE %s (%s)", unique.Alias, strings.Join(items, ", "))
	} else {
		sql += fmt.Sprintf("UNIQUE (%s)", strings.Join(items, ", "))
	}
	return sql
}

func (t *MySQLOperator) outputFullTextSQL(fulltext *model.FullText) string {
	var sql string
	items := make([]string, 0, len(fulltext.IndexColumns))
	for _, column := range fulltext.IndexColumns {
		items = append(items, fmt.Sprintf("`%s`", column.Name))
	}
	if fulltext.Alias != "" {
		sql += fmt.Sprintf("FULLTEXT %s (%s) WITH PARSER NGRAM", fulltext.Alias, strings.Join(items, ", "))
	} else {
		sql += fmt.Sprintf("FULLTEXT (%s) WITH PARSER NGRAM", strings.Join(items, ", "))
	}
	return sql
}

func (t *MySQLOperator) outputSingleIndexSQL(colName, indexAlias string) string {
	var sql string
	if indexAlias != "" {
		sql += fmt.Sprintf("INDEX %s (`%s`)", indexAlias, colName)
	} else {
		sql += fmt.Sprintf("INDEX (`%s`)", colName)
	}
	return sql
}

func (t *MySQLOperator) outputSingleUniqueSQL(colName, indexAlias string) string {
	var sql string
	if indexAlias != "" {
		sql += fmt.Sprintf("UNIQUE %s (`%s`)", indexAlias, colName)
	} else {
		sql += fmt.Sprintf("UNIQUE (`%s`)", colName)
	}
	return sql
}

func (t *MySQLOperator) outputSingleFullTextSQL(colName, indexAlias string) string {
	var sql string
	if indexAlias != "" {
		sql += fmt.Sprintf("FULLTEXT %s (`%s`) WITH PARSER NGRAM", indexAlias, colName)
	} else {
		sql += fmt.Sprintf("FULLTEXT (`%s`) WITH PARSER NGRAM", colName)
	}
	return sql
}

func (t *MySQLOperator) outputForeignKeySQL(foreignKey *model.ForeignKey) string {
	var sql string
	if foreignKey.Alias != "" {
		sql += fmt.Sprintf("CONSTRAINT %s FOREIGN KEY (`%s`) REFERENCES `%s` (`%s`)", foreignKey.Alias, foreignKey.Column.Name, foreignKey.Table, foreignKey.Primary)
	} else {
		sql += fmt.Sprintf("FOREIGN KEY (`%s`) REFERENCES `%s` (`%s`)", foreignKey.Column.Name, foreignKey.Table, foreignKey.Primary)
	}
	return sql
}
