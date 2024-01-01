package mysqloperator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	cerror "meta-egg/internal/error"
	"meta-egg/internal/model"

	jgmyql "github.com/Jinglever/go-mysql"
	log "github.com/sirupsen/logrus"
)

func (t *MySQLOperator) GetCurDBSchema() (*model.Database, error) {
	if t.DB == nil {
		return nil, cerror.ErrDBNotConnected
	}
	if t.CurDBSchema != nil {
		return t.CurDBSchema, nil
	}

	dbSchema := model.Database{
		Name:         t.DBName,
		Type:         model.DBType_MYSQL,
		IsSchemaByDB: true,
	}

	dbhelper := jgmyql.NewHelper(t.DB)

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
	r1, _ := regexp.Compile("CONSTRAINT `(.*)` FOREIGN KEY \\(`(.*)`\\) REFERENCES `(.*)` \\(`(.*)`\\)")
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
			// fmt.Println(line)
			var p, q int
			switch {
			case line[0] == '`': // column
				column := model.Column{}
				line = line[1:] // remove "`"

				// name
				p = strings.Index(line, "`")
				column.Name = line[:p]
				line = line[p+2:] // remove "...` "

				// type
				p = strings.Index(line, "(")
				q = strings.Index(line, " ")
				if q == -1 {
					q = len(line) - 1
				}
				if p == -1 || q < p {
					column.Type = model.ColumnType(strings.ToUpper(line[:q]))
				} else {
					column.Type = model.ColumnType(strings.ToUpper(line[:p]))
					p1 := strings.Index(line, ",")
					if p1 == -1 || p1 >= q {
						column.Length, _ = strconv.Atoi(line[p+1 : q-1])
					} else {
						column.Length, _ = strconv.Atoi(line[p+1 : p1])
						column.Decimal, _ = strconv.Atoi(line[p1+1 : q-1])
					}
				}
				if !model.AvailableColumnTypes[column.Type] {
					log.Errorf("unsupported column type %s", column.Type)
					return nil, cerror.ErrUnsupportedColumnType
				}
				line = line[q+1:]

				// unsigned
				p = strings.Index(line, "unsigned")
				if p != -1 {
					column.IsUnsigned = true
				}

				// NOT NULL
				p = strings.Index(line, "NOT NULL")
				if p != -1 {
					column.IsRequired = true
				}

				// DEFAULT
				p = strings.Index(line, "DEFAULT")
				if p != -1 && !strings.Contains(line, "DEFAULT NULL") {
					q = strings.LastIndex(line, "COMMENT")
					if q != -1 {
						column.InsertDefault = strings.Trim(line[p+8:q-1], "'")
					} else {
						column.InsertDefault = strings.Trim(line[p+8:len(line)-1], "'")
					}
					if column.Type == model.ColumnType_BINARY {
						// trim \0
						column.InsertDefault = strings.TrimRight(column.InsertDefault, "\\0")
					}
					if column.InsertDefault == "" && model.DefaultEmptyColTypes[column.Type] {
						column.DefaultEmpty = true
					}
				} else {
					column.InsertDefault = ""
				}

				// COMMENT
				p = strings.Index(line, "COMMENT")
				if p != -1 {
					column.Comment = line[p+9 : len(line)-2]
				}

				tableSchema.Columns = append(tableSchema.Columns, &column)
				colNameToColumn[column.Name] = &column
			case strings.HasPrefix(line, "PRIMARY KEY"): // primary key
				p = strings.Index(line, "`")
				q = strings.Index(line[p+1:], "`")
				colName := line[p+1 : p+q+1]
				for _, col := range tableSchema.Columns {
					if col.Name == colName {
						col.IsPrimaryKey = true
						break
					}
				}
			case strings.HasPrefix(line, "FULLTEXT KEY"): // fulltext key
				// must ngram
				if strings.Contains(line, "WITH PARSER `ngram`") {
					p = strings.Index(line, "`")
					q = strings.Index(line[p+1:], "`")
					alias := line[p+1 : p+q+1]

					p = strings.Index(line, "(`")
					q = strings.Index(line[p+2:], "`)")
					if q == -1 {
						q = strings.Index(line[p+2:], "`(") // for case: KEY `idx` (`id`(10))
					}
					colName := line[p+2 : p+q+2]
					if strings.Contains(colName, "`,`") {
						// 联合索引
						colNames := strings.Split(colName, "`,`")
						indexColumns := make([]*model.IndexColumn, 0, len(colNames))
						for _, col := range colNames {
							indexColumns = append(indexColumns, &model.IndexColumn{Name: col})
						}
						tableSchema.FullText = append(tableSchema.FullText, &model.FullText{IndexColumns: indexColumns, Alias: alias})
					} else {
						for _, col := range tableSchema.Columns {
							if col.Name == colName {
								col.IsFullText = true
								col.IndexAlias = alias
								break
							}
						}
					}
				}
			case strings.HasPrefix(line, "UNIQUE KEY"): // unique key
				p = strings.Index(line, "`")
				q = strings.Index(line[p+1:], "`")
				alias := line[p+1 : p+q+1]

				p = strings.Index(line, "(`")
				q = strings.Index(line[p+2:], "`)")
				if q == -1 {
					q = strings.Index(line[p+2:], "`(") // for case: KEY `idx` (`id`(10))
				}
				colName := line[p+2 : p+q+2]
				if strings.Contains(colName, "`,`") {
					// 联合索引
					colNames := strings.Split(colName, "`,`")
					indexColumns := make([]*model.IndexColumn, 0, len(colNames))
					for _, col := range colNames {
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
			case strings.HasPrefix(line, "KEY"): // index
				p = strings.Index(line, "`")
				q = strings.Index(line[p+1:], "`")
				alias := line[p+1 : p+q+1]

				p = strings.Index(line, "(`")
				q = strings.Index(line[p+2:], "`)")
				if q == -1 {
					q = strings.Index(line[p+2:], "`(") // for case: KEY `idx` (`id`(10))
				}
				colName := line[p+2 : p+q+2]
				if strings.Contains(colName, "`,`") {
					// 联合索引
					colNames := strings.Split(colName, "`,`")
					indexColumns := make([]*model.IndexColumn, 0, len(colNames))
					for _, col := range colNames {

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
				matches := r1.FindStringSubmatch(line)
				if len(matches) != 5 {
					log.Errorf("parse foreign key error: %s", line)
					return nil, fmt.Errorf("parse foreign key error: %s", line)
				}
				colNameToColumn[matches[2]].ForeignKeys = []*model.ForeignKey{{
					Table:   matches[3],
					Primary: matches[4],
					Alias:   matches[1],
				}}
			case strings.HasPrefix(line, ") ENGINE"): // engine
				// charset
				p = strings.Index(line, "DEFAULT CHARSET=")
				if p != -1 {
					q = strings.Index(line[p+16:], " ")
					if q == -1 {
						tableSchema.Charset = line[p+16:]
					} else {
						tableSchema.Charset = line[p+16 : p+16+q]
					}
				}

				// collate
				p = strings.Index(line, "COLLATE=")
				if p != -1 {
					q = strings.Index(line[p+8:], " ")
					if q == -1 {
						tableSchema.Collate = line[p+8:]
					} else {
						tableSchema.Collate = line[p+8 : p+8+q]
					}
				}

				// comment
				p = strings.Index(line, "COMMENT='")
				if p != -1 {
					q = strings.LastIndex(line, "'")
					if q != -1 {
						tableSchema.Comment = line[p+9 : q]
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
