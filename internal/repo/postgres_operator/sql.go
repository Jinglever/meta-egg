package pgoperator

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"meta-egg/internal/model"

	log "github.com/sirupsen/logrus"
)

func findAvailableTable(existsTableName map[string]bool,
	db *model.Database,
	t *model.Table, checkedTable map[string]bool) *model.Table {
	if checkedTable[t.Name] { // 防止循环依赖导致死循环
		return t
	}
	checkedTable[t.Name] = true
	for _, col := range t.Columns {
		if len(col.ForeignKeys) > 0 &&
			col.ForeignKeys[0].Table != t.Name && // 外链到自己的不算
			!existsTableName[col.ForeignKeys[0].Table] {
			return findAvailableTable(existsTableName, db, db.TableNameToTable[col.ForeignKeys[0].Table], checkedTable)
		}
	}
	return t
}

/*
 * 输出更新Schema的SQL
 * @param targetDBSchema 目标数据库的schema
 * @param createSQLWriter 输出创建数据库的SQL
 * @param incSQLWriter 输出更新数据库的增量SQL
 * @param metaDataSQLWriter 输出数据库的meta表的数据的INSERT SQL
 */
func (t *PostgreSQLOperator) OutputSQLForSchemaUpdating(targetDBSchema *model.Database,
	createSQLWriter io.StringWriter,
	incSQLWriter io.StringWriter,
	metaDataSQLWriter io.StringWriter,
) error {
	var wg sync.WaitGroup

	// create SQL
	wg.Add(1)
	go func() {
		defer wg.Done()
		createSQLWriter.WriteString(fmt.Sprintf("-- CREATE DATABASE \"%s\" WITH ENCODING '%s';\n",
			targetDBSchema.Name, targetDBSchema.Charset))
		createSQLWriter.WriteString(fmt.Sprintf("-- ALTER DATABASE \"%s\" OWNER to dev;\n",
			targetDBSchema.Name))
		var sortedTables []*model.Table
		existsTableName := make(map[string]bool)
		for _, table := range targetDBSchema.Tables {
			if existsTableName[table.Name] {
				continue
			}
			for {
				// 也许不算最高效的算法，但是足够简单易懂，且实际上表的数量不会很多
				t := findAvailableTable(existsTableName, targetDBSchema, table, make(map[string]bool))
				sortedTables = append(sortedTables, t)
				existsTableName[t.Name] = true
				if t == table {
					break
				}
			}
		}
		for _, table := range sortedTables {
			t.outputCreateTableSQL(table, createSQLWriter)
		}
	}()

	// inc SQL
	wg.Add(1)
	go func() {
		defer wg.Done()
		curDBSchema, err := t.GetCurDBSchema()
		if err != nil {
			log.Errorf("get current db schema failed: %v", err)
			return
		}

		// database的以下属性只能手工变更：name、charset、collate
		// if curDBSchema.Name != targetDBSchema.Name {
		// 	log.Warnf("need modify manually: database.name")
		// }
		if curDBSchema.Charset != targetDBSchema.Charset {
			log.Warnf("need modify manually: database.charset")
		}
		// 这里比较复杂，暂时不支持，用默认的吧
		// if curDBSchema.Collate != targetDBSchema.Collate {
		// 	log.Warnf("need modify manually: database.collate")
		// }

		// tables
		// todo table.name modify dispend on dtd rule support
		// todo column.name modify dispend on dtd rule support

		existsTableName := make(map[string]bool)
		for _, table := range curDBSchema.Tables {
			if _, ok := targetDBSchema.TableNameToTable[table.Name]; !ok {
				// rm tables
				t.outputRemoveTableSQL(table, incSQLWriter)
			} else {
				existsTableName[table.Name] = true
			}
		}
		// add tables
		var sortedNeedAddTables []*model.Table
		for _, table := range targetDBSchema.Tables {
			if existsTableName[table.Name] {
				continue
			}
			for {
				// 也许不算最高效的算法，但是足够简单易懂，且实际上表的数量不会很多
				t := findAvailableTable(existsTableName, targetDBSchema, table, make(map[string]bool))
				sortedNeedAddTables = append(sortedNeedAddTables, t)
				existsTableName[t.Name] = true
				if t == table {
					break
				}
			}
		}
		for _, table := range sortedNeedAddTables {
			t.outputCreateTableSQL(table, incSQLWriter)
		}
		// modify
		for _, table := range targetDBSchema.Tables {
			if _, ok := curDBSchema.TableNameToTable[table.Name]; ok {
				t.outputModifyTableSQL(curDBSchema.TableNameToTable[table.Name], table, incSQLWriter)
			}
		}
	}()

	// meta data sql
	wg.Add(1)
	go func() {
		defer wg.Done()
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
			metaDataSQLWriter.WriteString(fmt.Sprintf("INSERT INTO \"%s\" (\"%s\") VALUES\n",
				table.Name, strings.Join(colNames, "\",\"")))
			for i, record := range allRecords {
				metaDataSQLWriter.WriteString("(")
				for j, colName := range colNames {
					if record[colName] == nil {
						metaDataSQLWriter.WriteString("NULL")
					} else {
						if v, ok := record[colName].(string); ok {
							metaDataSQLWriter.WriteString(fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''")))
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

func (t *PostgreSQLOperator) outputCreateTableSQL(table *model.Table, writer io.StringWriter) {
	// create auto increment sequence
	writer.WriteString(fmt.Sprintf("\nCREATE SEQUENCE IF NOT EXISTS %s_%s_seq;\n",
		table.Name, table.PrimaryColumn.Name))
	// create table
	writer.WriteString(fmt.Sprintf("CREATE TABLE \"%s\" (\n", table.Name))

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

	// foreign key
	for _, column := range table.Columns {
		if len(column.ForeignKeys) > 0 {
			writer.WriteString(",\n")
			writer.WriteString(fmt.Sprintf("  %s", t.outputForeignKeySQL(column.ForeignKeys[0])))
		}
	}

	// table comment
	writer.WriteString("\n);")

	// single index
	for _, column := range table.Columns {
		if column.IsUnique && len(column.ForeignKeys) == 0 {
			writer.WriteString("\n")
			writer.WriteString(t.outputSingleUniqueSQL(table.Name, column.Name, column.IndexAlias))
		}
		if column.IsIndex && len(column.ForeignKeys) == 0 {
			writer.WriteString("\n")
			writer.WriteString(t.outputSingleIndexSQL(table.Name, column.Name, column.IndexAlias))
		}
	}

	// multiple unique index
	for _, unique := range table.Unique {
		writer.WriteString("\n")
		writer.WriteString(t.outputUniqueSQL(table.Name, unique))
	}

	// multiple index
	for _, index := range table.Indexes {
		writer.WriteString("\n")
		writer.WriteString(t.outputIndexSQL(table.Name, index))
	}

	// comment
	if table.Comment != "" {
		writer.WriteString("\n")
		writer.WriteString(fmt.Sprintf("COMMENT ON TABLE \"%s\" IS '%s';", table.Name, table.Comment))
	}
	for _, col := range table.Columns {
		if col.Comment != "" {
			writer.WriteString("\n")
			writer.WriteString(fmt.Sprintf("COMMENT ON COLUMN \"%s\".\"%s\" IS '%s';", table.Name, col.Name, col.Comment))
		}
	}
	writer.WriteString("\n")
}

func (t *PostgreSQLOperator) outputRemoveTableSQL(table *model.Table, writer io.StringWriter) {
	if !t.IgnoreTables[table.Name] {
		writer.WriteString(fmt.Sprintf("DROP TABLE \"%s\";\n", table.Name))
	}
}

func (t *PostgreSQLOperator) outputModifyTableSQL(curTable *model.Table, targetTable *model.Table, writer io.StringWriter) {
	if curTable.Comment != targetTable.Comment {
		writer.WriteString(fmt.Sprintf("COMMENT ON TABLE public.\"%s\" IS '%s';\n", targetTable.Name, targetTable.Comment))
	}
	// multi-index
	for key, idx := range curTable.ColNamesToIndex {
		if _, ok := targetTable.ColNamesToIndex[key]; !ok {
			// rm index
			if idx.Alias != "" {
				writer.WriteString(fmt.Sprintf("DROP INDEX public.%s;\n", idx.Alias))
			}
		}
	}
	for key, idx := range curTable.ColNamesToUnique {
		if _, ok := targetTable.ColNamesToUnique[key]; !ok {
			// rm unique
			if idx.Alias != "" {
				writer.WriteString(fmt.Sprintf("DROP INDEX public.%s;\n", idx.Alias))
			}
		}
	}
	// column
	for _, column := range curTable.ColNameToColumn {
		if _, ok := targetTable.ColNameToColumn[column.Name]; !ok {
			// rm index & foreign key first if exists
			// drop foreign key first
			if len(column.ForeignKeys) > 0 && column.ForeignKeys[0].Alias != "" {
				writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" DROP CONSTRAINT %s;\n", targetTable.Name, column.ForeignKeys[0].Alias))
			}
			if column.IsIndex && column.IndexAlias != "" {
				writer.WriteString(fmt.Sprintf("DROP INDEX %s;\n", column.IndexAlias))
			}
			if column.IsUnique && column.IndexAlias != "" {
				writer.WriteString(fmt.Sprintf("DROP INDEX %s;\n", column.IndexAlias))
			}
			// rm column
			writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" DROP COLUMN \"%s\";\n", targetTable.Name, column.Name))
		}
	}
	for _, column := range targetTable.ColNameToColumn {
		curCol, ok := curTable.ColNameToColumn[column.Name]
		if !ok {
			// add column
			// ALTER TABLE "public"."baseline" ADD COLUMN "abc" timestamp;
			// COMMENT ON COLUMN "public"."baseline"."abc" IS 'abc';
			if column.IsPrimaryKey && model.IntegerColumnTypes[column.Type] {
				// create sequence
				writer.WriteString(fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS %s_%s_seq;\n",
					targetTable.Name, column.Name))
			}
			writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" ADD COLUMN %s;\n", targetTable.Name, t.outputColumnSQL(column)))
			if column.Comment != "" {
				writer.WriteString(fmt.Sprintf("COMMENT ON COLUMN public.\"%s\".\"%s\" IS '%s';\n", targetTable.Name, column.Name, column.Comment))
			}
			if column.IsUnique {
				writer.WriteString(fmt.Sprintf("%s\n", t.outputSingleUniqueSQL(targetTable.Name, column.Name, column.IndexAlias)))
			}
			if column.IsIndex {
				writer.WriteString(fmt.Sprintf("%s\n", t.outputSingleIndexSQL(targetTable.Name, column.Name, column.IndexAlias)))
			}
			if len(column.ForeignKeys) > 0 {
				// ALTER TABLE "public"."baseline" ADD FOREIGN KEY ("load_id") REFERENCES "public"."load" ("id");
				writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" ADD %s;\n", targetTable.Name, t.outputForeignKeySQL(column.ForeignKeys[0])))
			}
		} else {
			modifyByDropThenAdd := false
			if column.Type != curCol.Type {
				if column.Type == model.ColumnType_BOOL {
					// drop first then add new one
					writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" DROP COLUMN \"%s\";\n", targetTable.Name, column.Name))
					writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" ADD COLUMN %s;\n", targetTable.Name, t.outputColumnSQL(column)))
					modifyByDropThenAdd = true
				} else {
					// ALTER TABLE table_name ALTER COLUMN column_name TYPE new_data_type;
					writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" ALTER COLUMN \"%s\" TYPE %s;\n", targetTable.Name, column.Name, GetPgType(column)))
				}
			} else if (column.Length != curCol.Length &&
				(model.StringColumnTypes[column.Type] ||
					column.Type == model.ColumnType_DECIMAL ||
					model.BinaryColumnTypes[column.Type])) ||
				column.Decimal != curCol.Decimal {
				// ALTER TABLE table_name ALTER COLUMN column_name TYPE new_data_type;
				writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" ALTER COLUMN \"%s\" TYPE %s;\n", targetTable.Name, column.Name, GetPgType(column)))
			}
			if column.IsRequired != curCol.IsRequired && !modifyByDropThenAdd {
				// ALTER TABLE table_name ALTER COLUMN column_name SET NOT NULL;
				// ALTER TABLE table_name ALTER COLUMN column_name DROP NOT NULL;
				if column.IsRequired {
					writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" ALTER COLUMN \"%s\" SET NOT NULL;\n", targetTable.Name, column.Name))
				} else {
					writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" ALTER COLUMN \"%s\" DROP NOT NULL;\n", targetTable.Name, column.Name))
				}
			}
			if column.IsPrimaryKey != curCol.IsPrimaryKey {
				// ALTER TABLE table_name ADD PRIMARY KEY (column_list);
				// ALTER TABLE table_name DROP CONSTRAINT table_name_pkey;
				if column.IsPrimaryKey {
					writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" DROP CONSTRAINT %s_pkey;\n", targetTable.Name, targetTable.Name))
					writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" ADD PRIMARY KEY (\"%s\");\n", targetTable.Name, column.Name))
				}
			}
			if column.Comment != curCol.Comment {
				// COMMENT ON COLUMN "public"."baseline"."abc" IS 'abc';
				writer.WriteString(fmt.Sprintf("COMMENT ON COLUMN public.\"%s\".\"%s\" IS '%s';\n", targetTable.Name, column.Name, column.Comment))
			}
			if ((column.DefaultEmpty != curCol.DefaultEmpty &&
				model.DefaultEmptyColTypes[column.Type]) ||
				(column.InsertDefault != curCol.InsertDefault &&
					!model.NoDefaultColTypes[column.Type] &&
					!(model.DefaultEmptyColTypes[column.Type] &&
						column.DefaultEmpty))) &&
				!modifyByDropThenAdd {
				// ALTER TABLE table_name ALTER COLUMN column_name SET DEFAULT expression;
				// ALTER TABLE table_name ALTER COLUMN column_name DROP DEFAULT;
				if (column.DefaultEmpty && model.DefaultEmptyColTypes[column.Type]) ||
					column.InsertDefault != "" {
					if model.StringColumnTypes[column.Type] ||
						model.BinaryColumnTypes[column.Type] {
						writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" ALTER COLUMN \"%s\" SET DEFAULT '%s';\n", targetTable.Name, column.Name, column.InsertDefault))
					} else {
						if column.IsPrimaryKey {
							writer.WriteString(fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS %s_%s_seq;\n",
								targetTable.Name, column.Name))
						}
						writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" ALTER COLUMN \"%s\" SET DEFAULT %s;\n", targetTable.Name, column.Name, column.InsertDefault))
					}
				} else {
					writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" ALTER COLUMN \"%s\" DROP DEFAULT;\n", targetTable.Name, column.Name))
				}
			}
			// single-index
			if curCol.IsIndex != column.IsIndex {
				if column.IsIndex {
					writer.WriteString(fmt.Sprintf("%s\n", t.outputSingleIndexSQL(targetTable.Name, column.Name, column.IndexAlias)))
				} else if curCol.IndexAlias != "" {
					writer.WriteString(fmt.Sprintf("DROP INDEX %s;\n", curCol.IndexAlias))
				}
			}
			if curCol.IsUnique != column.IsUnique {
				if column.IsUnique {
					writer.WriteString(fmt.Sprintf("%s\n", t.outputSingleUniqueSQL(targetTable.Name, column.Name, column.IndexAlias)))
				} else if curCol.IndexAlias != "" {
					writer.WriteString(fmt.Sprintf("DROP INDEX %s;\n", curCol.IndexAlias))
				}
			}
			// foreign key
			if len(curCol.ForeignKeys) != len(column.ForeignKeys) {
				if len(column.ForeignKeys) > 0 {
					writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" ADD %s;\n", targetTable.Name, t.outputForeignKeySQL(column.ForeignKeys[0])))
				} else if len(curCol.ForeignKeys) > 0 && curCol.ForeignKeys[0].Alias != "" {
					writer.WriteString(fmt.Sprintf("ALTER TABLE public.\"%s\" DROP CONSTRAINT %s;\n", targetTable.Name, curCol.ForeignKeys[0].Alias))
				}
			}
		}
	}
	// add multi-index
	for key, idx := range targetTable.ColNamesToIndex {
		if _, ok := curTable.ColNamesToIndex[key]; !ok {
			// add index
			writer.WriteString(fmt.Sprintf("%s\n", t.outputIndexSQL(targetTable.Name, idx)))
		}
	}
	for key, idx := range targetTable.ColNamesToUnique {
		if _, ok := curTable.ColNamesToUnique[key]; !ok {
			// add unique
			writer.WriteString(fmt.Sprintf("%s\n", t.outputUniqueSQL(targetTable.Name, idx)))
		}
	}
}

func (t *PostgreSQLOperator) outputColumnSQL(column *model.Column) string {
	var sql string

	if column == column.Table.PrimaryColumn {
		// primary key
		sql += `"` + column.Name + `"` + " "
		sql += GetPgType(column)
		sql += " PRIMARY KEY"
		sql += fmt.Sprintf(" DEFAULT nextval('%s_%s_seq')", column.Table.Name, column.Name)

	} else {
		// normal column
		sql += `"` + column.Name + `"` + " "
		sql += GetPgType(column)
		if column.IsRequired {
			sql += " NOT NULL"
		}
		if !model.NoDefaultColTypes[column.Type] {
			if column.DefaultEmpty && model.DefaultEmptyColTypes[column.Type] {
				sql += " DEFAULT ''"
			} else if column.InsertDefault != "" {
				if model.StringColumnTypes[column.Type] || model.BinaryColumnTypes[column.Type] {
					sql += fmt.Sprintf(" DEFAULT '%s'", column.InsertDefault)
				} else {
					sql += fmt.Sprintf(" DEFAULT %s", column.InsertDefault)
				}
			}
		}
	}

	return sql
}

func (t *PostgreSQLOperator) outputIndexSQL(tableName string, index *model.Index) string {
	var sql string
	items := make([]string, 0, len(index.IndexColumns))
	for _, column := range index.IndexColumns {
		items = append(items, column.Name)
	}
	if index.Alias != "" {
		sql += fmt.Sprintf("CREATE INDEX %s ON public.\"%s\" (\"%s\");", index.Alias, tableName, strings.Join(items, "\", \""))
	} else {
		sql += fmt.Sprintf("CREATE INDEX ON public.\"%s\" (\"%s\");", tableName, strings.Join(items, "\", \""))
	}
	return sql
}

func (t *PostgreSQLOperator) outputUniqueSQL(tableName string, unique *model.Unique) string {
	var sql string
	items := make([]string, 0, len(unique.IndexColumns))
	for _, column := range unique.IndexColumns {
		items = append(items, column.Name)
	}
	if unique.Alias != "" {
		sql += fmt.Sprintf("CREATE UNIQUE INDEX %s ON public.\"%s\" (\"%s\");", unique.Alias, tableName, strings.Join(items, "\", \""))
	} else {
		sql += fmt.Sprintf("CREATE UNIQUE INDEX ON public.\"%s\" (\"%s\");", tableName, strings.Join(items, "\", \""))
	}
	return sql
}

func (t *PostgreSQLOperator) outputSingleIndexSQL(tableName, colName, indexAlias string) string {
	var sql string
	if indexAlias != "" {
		sql += fmt.Sprintf("CREATE INDEX %s ON public.\"%s\" (\"%s\");", indexAlias, tableName, colName)
	} else {
		sql += fmt.Sprintf("CREATE INDEX ON public.\"%s\" (\"%s\");", tableName, colName)
	}
	return sql
}

func (t *PostgreSQLOperator) outputSingleUniqueSQL(tableName, colName, indexAlias string) string {
	var sql string
	if indexAlias != "" {
		sql += fmt.Sprintf("CREATE UNIQUE INDEX %s ON public.\"%s\" (\"%s\");", indexAlias, tableName, colName)
	} else {
		sql += fmt.Sprintf("CREATE UNIQUE INDEX ON public.\"%s\" (\"%s\");", tableName, colName)
	}
	return sql
}

func (t *PostgreSQLOperator) outputForeignKeySQL(foreignKey *model.ForeignKey) string {
	var sql string
	if foreignKey.Alias != "" {
		sql += fmt.Sprintf("CONSTRAINT %s FOREIGN KEY (\"%s\") REFERENCES \"%s\"(\"%s\")", foreignKey.Alias, foreignKey.Column.Name, foreignKey.Table, foreignKey.Primary)
	} else {
		sql += fmt.Sprintf("FOREIGN KEY (\"%s\") REFERENCES \"%s\"(\"%s\")", foreignKey.Column.Name, foreignKey.Table, foreignKey.Primary)
	}
	return sql
}
