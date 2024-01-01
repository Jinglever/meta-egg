package pgoperator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	cerror "meta-egg/internal/error"
	"meta-egg/internal/model"

	jgpg "github.com/Jinglever/go-postgres"
	log "github.com/sirupsen/logrus"
)

func (t *PostgreSQLOperator) GetCurDBSchema() (*model.Database, error) {
	if t.DB == nil {
		return nil, cerror.ErrDBNotConnected
	}
	if t.CurDBSchema != nil {
		return t.CurDBSchema, nil
	}

	dbSchema := model.Database{
		Name:         t.DBName,
		Type:         model.DBType_PG,
		IsSchemaByDB: true,
	}

	dbhelper := jgpg.NewHelper(t.DB)

	// get database charset
	dbCharset, err := dbhelper.QueryDBCharset()
	if err != nil {
		log.Errorf("get database charset failed, err: %v", err)
		return nil, err
	}
	dbSchema.Charset = dbCharset

	// get database collate
	dbCollate, err := dbhelper.QueryDBCollate()
	if err != nil {
		log.Errorf("get database collate failed, err: %v", err)
		return nil, err
	}
	dbSchema.Collate = dbCollate

	// get all tables
	tables, err := dbhelper.QueryAllTables()
	if err != nil {
		log.Errorf("get all tables failed, err: %v", err)
		return nil, err
	}
	log.Debugf("database has %d tables", len(tables))
	dbSchema.Tables = make([]*model.Table, 0, len(tables))

	// get all columns for each table
	r1, _ := regexp.Compile(`CONSTRAINT (.*) FOREIGN KEY \((.*)\) REFERENCES (.*)\((.*)\)`)
	for _, table := range tables {
		tableSchema := model.Table{
			Name: table,
		}

		rawSql, err := dbhelper.QueryCreateTableSql(table)
		if err != nil {
			log.Errorf("get all create table sql for table %s failed, err: %v", table, err)
			return nil, err
		}
		log.Debugf("table %s create sql: %s", table, rawSql)

		// split lines
		lines := strings.Split(rawSql, "\n")

		colNameToColumn := make(map[string]*model.Column)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var p, q int
			switch {
			case line[0] != 'C' && line[0] != ')': // column
				column := model.Column{}

				// name
				p = strings.Index(line, " ")
				column.Name = line[:p]
				column.Name = strings.Trim(column.Name, "\"")
				line = line[p+1:] // remove "... "

				// type
				p = strings.Index(line, "(")
				q = strings.Index(line, " ")
				if p == -1 || (q != -1 && q < p) {
					if q == -1 {
						line = strings.TrimSpace(line)
						line = strings.Trim(line, ",")
						column.Type = ConvertPgType2ModelColType(strings.ToUpper(line))
					} else {
						column.Type = ConvertPgType2ModelColType(strings.ToUpper(line[:q]))
					}
				} else {
					column.Type = ConvertPgType2ModelColType(strings.ToUpper(line[:p]))
					q = strings.Index(line, ")")
					p1 := strings.Index(line[:q], ",")
					if p1 == -1 || p1 > q {
						column.Length, _ = strconv.Atoi(line[p+1 : q])
					} else {
						column.Length, _ = strconv.Atoi(line[p+1 : p1])
						column.Decimal, _ = strconv.Atoi(line[p1+1 : q])
					}
				}
				if !model.AvailableColumnTypes[column.Type] {
					log.Errorf("unsupported column type %s", column.Type)
					return nil, cerror.ErrUnsupportedColumnType
				}
				line = line[q+1:]

				// PRIMARY KEY
				p = strings.Index(line, "PRIMARY KEY")
				if p != -1 {
					column.IsPrimaryKey = true
				}

				// NOT NULL
				p = strings.Index(line, "NOT NULL")
				if p != -1 {
					column.IsRequired = true
				}

				// DEFAULT
				p = strings.Index(line, "DEFAULT")
				if p != -1 && !strings.Contains(line, "DEFAULT NULL") {
					column.InsertDefault = strings.Trim(line[p+8:], ",")
					if strings.Contains(column.InsertDefault, "::") {
						p = strings.Index(column.InsertDefault, "::")
						q = strings.Index(column.InsertDefault[p+2:], ")")
						tmpStr := column.InsertDefault[:p]
						if q != -1 {
							tmpStr += column.InsertDefault[p+2+q:]
						}
						column.InsertDefault = tmpStr
					}
					column.InsertDefault = strings.Trim(column.InsertDefault, "'")
					column.InsertDefault = strings.Trim(column.InsertDefault, "\"")
					if column.Type == model.ColumnType_TINYINT {
						if strings.ToLower(column.InsertDefault) == "true" {
							column.InsertDefault = "1"
						} else if strings.ToLower(column.InsertDefault) == "false" {
							column.InsertDefault = "0"
						}
					}
					if column.InsertDefault == "" && model.DefaultEmptyColTypes[column.Type] {
						column.DefaultEmpty = true
					}
				} else {
					column.InsertDefault = ""
				}

				tableSchema.Columns = append(tableSchema.Columns, &column)
				colNameToColumn[column.Name] = &column
			case strings.HasPrefix(line, "CREATE UNIQUE INDEX"): // unique key
				// example: CREATE UNIQUE INDEX baseline_respond_date_idx ON public.baseline USING btree (respond_date);
				p = len("CREATE UNIQUE INDEX")
				q = strings.Index(line[p+1:], " ON")
				alias := line[p+1 : p+q+1]

				p = strings.Index(line, "(")
				q = strings.Index(line[p+1:], ")")
				colName := strings.TrimSpace(line[p+1 : p+q+1])
				colName = strings.Trim(colName, "\"")
				if strings.Contains(colName, ",") {
					// 联合索引
					colNames := strings.Split(colName, ",")
					indexColumns := make([]*model.IndexColumn, 0, len(colNames))
					for _, col := range colNames {
						col = strings.Trim(strings.TrimSpace(col), "\"")
						indexColumns = append(indexColumns, &model.IndexColumn{Name: col})
					}
					tableSchema.Unique = append(tableSchema.Unique, &model.Unique{IndexColumns: indexColumns, Alias: alias})
				} else {
					for _, col := range tableSchema.Columns {
						if col.Name == colName {
							col.IsUnique = true
							col.IndexAlias = alias
							break
						}
					}
				}
			case strings.HasPrefix(line, "CREATE INDEX"): // index
				// example: CREATE INDEX baseline_deleted_at_idx ON public.baseline USING btree (deleted_at);
				p = len("CREATE INDEX")
				q = strings.Index(line[p+1:], " ON")
				alias := line[p+1 : p+q+1]

				p = strings.Index(line, "(")
				q = strings.Index(line[p+1:], ")")
				colName := strings.TrimSpace(line[p+1 : p+q+1])
				colName = strings.Trim(colName, "\"")
				if strings.Contains(colName, ",") {
					// 联合索引
					colNames := strings.Split(colName, ",")
					indexColumns := make([]*model.IndexColumn, 0, len(colNames))
					for _, col := range colNames {
						col = strings.Trim(strings.TrimSpace(col), "\"")
						indexColumns = append(indexColumns, &model.IndexColumn{Name: col})
					}
					tableSchema.Indexes = append(tableSchema.Indexes, &model.Index{IndexColumns: indexColumns, Alias: alias})
				} else {
					for _, col := range tableSchema.Columns {
						if col.Name == colName {
							col.IsIndex = true
							col.IndexAlias = alias
							break
						}
					}
				}
			case strings.HasPrefix(line, "CONSTRAINT"): // foreign key
				// example: CONSTRAINT baseline_load_id_fkey FOREIGN KEY (load_id) REFERENCES load(id)
				matches := r1.FindStringSubmatch(line)
				if len(matches) != 5 {
					log.Errorf("parse foreign key error: %s", line)
					return nil, fmt.Errorf("parse foreign key error: %s", line)
				}
				colNameToColumn[matches[2]].ForeignKeys = []*model.ForeignKey{{
					Table:   strings.Trim(matches[3], "\""),
					Primary: strings.Trim(matches[4], "\""),
					Alias:   matches[1],
				}}
			case strings.HasPrefix(line, "COMMENT ON TABLE"): // comment for table
				// example: COMMENT ON TABLE baseline IS '基线表';
				p = strings.Index(line, "IS")
				q = strings.Index(line[p+2:], ";")
				comment := strings.TrimSpace(line[p+2 : p+q+2])
				comment = strings.Trim(comment, "'")
				comment = strings.Trim(comment, "\"")
				tableSchema.Comment = comment
			case strings.HasPrefix(line, "COMMENT ON COLUMN"): // comment for column
				// example: COMMENT ON COLUMN baseline.load_id IS '所属负荷系统';
				p = strings.Index(line, ".")
				q = strings.Index(line[p+1:], " ")
				colName := line[p+1 : p+q+1]
				line = line[p+q+1:]

				p = strings.Index(line, "IS")
				q = strings.Index(line[p+2:], ";")
				comment := strings.TrimSpace(line[p+2 : p+q+2])
				comment = strings.Trim(comment, "'")
				comment = strings.Trim(comment, "\"")

				for _, col := range tableSchema.Columns {
					if col.Name == colName {
						col.Comment = comment
						break
					}
				}
			}
		}
		dbSchema.Tables = append(dbSchema.Tables, &tableSchema)
	}

	dbSchema.MakeUp(nil)

	t.CurDBSchema = &dbSchema
	return t.CurDBSchema, nil
}
