package model

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (p *Project) Validate() error {
	if p.Name == "" {
		log.Errorf("project name is empty")
		return fmt.Errorf("project name is empty")
	}
	if p.Database != nil {
		p.Database.Project = p
		err := p.Database.Validate()
		if err != nil {
			return err
		}
	}
	if p.Domain != nil {
		p.Domain.Project = p
		err := p.Domain.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Domain) Validate() error {
	for _, usecase := range d.Usecases {
		usecase.Domain = d
		err := usecase.Validate(d)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *Usecase) Validate(d *Domain) error {
	if u.Name == "" {
		log.Errorf("usecase name is empty")
		return fmt.Errorf("usecase name is empty")
	}
	if u.Name == "biz" || u.Name == "option" || u.Name == "usecase" {
		log.Errorf("usecase name (%s) is reserved", u.Name)
		return fmt.Errorf("usecase name (%s) is reserved", u.Name)
	}
	if u.Desc == "" {
		log.Errorf("usecase (%s) desc is empty", u.Name)
		return fmt.Errorf("usecase (%s) desc is empty", u.Name)
	}
	return nil
}

func (d *Database) Validate() error {
	if d.Name == "" {
		log.Errorf("database name is empty")
		return fmt.Errorf("database name is empty")
	}

	for _, table := range d.Tables {
		table.Database = d
		err := table.Validate(d)
		if err != nil {
			return err
		}
	}

	// validate foreign key reference table exists
	tableNameToTable := make(map[string]*Table)
	for _, table := range d.Tables {
		tableNameToTable[table.Name] = table
	}
	for _, table := range d.Tables {
		riForeignKeyCount := 0
		for _, column := range table.Columns {
			// 外链相关
			for _, foreignKey := range column.ForeignKeys {
				if _, ok := tableNameToTable[foreignKey.Table]; !ok {
					log.Errorf("foreign key reference table (%s) not exists", foreignKey.Table)
					return fmt.Errorf("foreign key reference table (%s) not exists", foreignKey.Table)
				}
				if foreignKey.Counter != "" {
					isColumnExists := false
					for _, c := range tableNameToTable[table.Name].Columns {
						if c.Name == foreignKey.Counter {
							if IntegerColumnTypes[c.Type] {
								isColumnExists = true
								break
							} else {
								log.Errorf("foreign key reference counter column (%s) type (%s) is not integer", foreignKey.Counter, c.Type)
								return fmt.Errorf("foreign key reference counter column (%s) type (%s) is not integer", foreignKey.Counter, c.Type)
							}
						}
					}
					if !isColumnExists {
						log.Errorf("foreign key reference counter column (%s) not exists", foreignKey.Counter)
						return fmt.Errorf("foreign key reference counter column (%s) not exists", foreignKey.Counter)
					}
				}
				if foreignKey.RiRaw != "" {
					isColumnExists := false
					for _, c := range tableNameToTable[foreignKey.Table].Columns {
						if c.Name == foreignKey.RiRaw {
							if StringColumnTypes[c.Type] {
								isColumnExists = true
								break
							} else {
								log.Errorf("foreign key reference riRaw column (%s) type (%s) is not string", foreignKey.RiRaw, c.Type)
								return fmt.Errorf("foreign key reference riRaw column (%s) type (%s) is not string", foreignKey.RiRaw, c.Type)
							}
						}
					}
					if !isColumnExists {
						log.Errorf("foreign key reference ri_raw column (%s) not exists", foreignKey.RiRaw)
						return fmt.Errorf("foreign key reference ri_raw column (%s) not exists", foreignKey.RiRaw)
					}

					riKeyCount := 0
					for _, c := range tableNameToTable[table.Name].Columns {
						if StringColumnTypes[c.Type] && c.IsRiKey {
							riKeyCount++
						}
					}
					if riKeyCount == 0 {
						log.Errorf("foreign key reference ri_raw column (%s) but ri_key column not exists", foreignKey.RiRaw)
						return fmt.Errorf("foreign key reference ri_raw column (%s) but ri_key column not exists", foreignKey.RiRaw)
					} else if riKeyCount > 1 {
						log.Errorf("foreign key reference ri_raw column (%s) but ri_key column count (%d) > 1", foreignKey.RiRaw, riKeyCount)
						return fmt.Errorf("foreign key reference ri_raw column (%s) but ri_key column count (%d) > 1", foreignKey.RiRaw, riKeyCount)
					}

					if table.Type != TableType_RI {
						log.Errorf("foreign key reference ri_raw column (%s) but table type (%s) is not ri", foreignKey.RiRaw, table.Type)
						return fmt.Errorf("foreign key reference ri_raw column (%s) but table type (%s) is not ri", foreignKey.RiRaw, table.Type)
					}

					riForeignKeyCount++
				}
			}

			// 联合索引相关
			for _, index := range table.Indexes {
				for _, indexColumn := range index.IndexColumns {
					isIndexColumnExists := false
					for _, column := range table.Columns {
						if column.Name == indexColumn.Name {
							isIndexColumnExists = true
							break
						}
					}
					if !isIndexColumnExists {
						log.Errorf("index column (%s) not exists", indexColumn.Name)
						return fmt.Errorf("index column (%s) not exists", indexColumn.Name)
					}
				}
			}

			// 联合唯一索引相关
			for _, index := range table.Unique {
				for _, indexColumn := range index.IndexColumns {
					isIndexColumnExists := false
					for _, column := range table.Columns {
						if column.Name == indexColumn.Name {
							isIndexColumnExists = true
							break
						}
					}
					if !isIndexColumnExists {
						log.Errorf("unique index column (%s) not exists", indexColumn.Name)
						return fmt.Errorf("unique index column (%s) not exists", indexColumn.Name)
					}
				}
			}

			// 联合全文搜索索引相关
			for _, index := range table.FullText {
				for _, indexColumn := range index.IndexColumns {
					isIndexColumnExists := false
					for _, column := range table.Columns {
						if column.Name == indexColumn.Name {
							isIndexColumnExists = true
							break
						}
					}
					if !isIndexColumnExists {
						log.Errorf("fulltext index column (%s) not exists", indexColumn.Name)
						return fmt.Errorf("fulltext index column (%s) not exists", indexColumn.Name)
					}
				}
			}
		}

		if table.Type == TableType_RI && riForeignKeyCount == 0 {
			log.Errorf("ri table (%s) has no ri foreign key", table.Name)
			return fmt.Errorf("ri table (%s) has no ri foreign key", table.Name)
		}

		// todo 检查外链和索引是否有重复
		// todo 检查联合索引和联合唯一索引是否有重复
	}
	return nil
}

func (t *Table) Validate(d *Database) error {
	if t.Name == "" {
		log.Errorf("table name is empty")
		return fmt.Errorf("table name is empty")
	}
	if strings.ToLower(t.Name) == "base" { // 保留词约束
		log.Errorf("table name (%s) is reserved", t.Name)
		return fmt.Errorf("table name (%s) is reserved", t.Name)
	}
	if !AvailableTableTypes[t.Type] {
		log.Errorf("table type (%s) is not available", t.Type)
		return fmt.Errorf("table type (%s) is not available", t.Type)
	}
	if t.Comment == "" { // 约束：表注释不能为空
		log.Errorf("table (%s) comment is empty", t.Name)
		return fmt.Errorf("table (%s) comment is empty", t.Name)
	}

	for _, column := range t.Columns {
		column.Table = t
		if err := column.Validate(t); err != nil {
			return err
		}
	}

	for _, index := range t.Indexes {
		index.Table = t
		err := index.Validate(t)
		if err != nil {
			return err
		}
	}

	for _, index := range t.Unique {
		index.Table = t
		err := index.Validate(t)
		if err != nil {
			return err
		}
	}

	for _, index := range t.FullText {
		index.Table = t
		err := index.Validate(t)
		if err != nil {
			return err
		}
	}

	if t.Type == TableType_META {
		isSemanticExists := false
		for _, column := range t.Columns {
			if column.Name == "semantic" ||
				column.Name == "sematic" { // 这个是为了兼容之前的写法
				isSemanticExists = true
				break
			}
		}
		if !isSemanticExists {
			log.Errorf("meta table (%s) has no semantic column", t.Name)
			return fmt.Errorf("meta table (%s) has no semantic column", t.Name)
		}
	}

	return nil
}

func (c *Column) Validate(t *Table) error {
	if c.Name == "" {
		log.Errorf("column name is empty")
		return fmt.Errorf("column name is empty")
	}
	if !AvailableColumnTypes[c.Type] {
		if _, ok := ColumnExtTypeToColumnType[ColumnExtType(c.Type)]; !ok {
			log.Errorf("column (%s) type (%s) is invalid", c.Name, c.Type)
			return fmt.Errorf("column (%s) type (%s) is invalid", c.Name, c.Type)
		}

		// if ColumnExtType(c.Type) == ColumnExtType_FID && len(c.ForeignKeys) == 0 {
		// 	log.Errorf("column (%s) extType (%s) must have foreign key", c.Name, c.Type)
		// 	return fmt.Errorf("column (%s) extType (%s) must have foreign key", c.Name, c.Type)
		// }

		if ColumnExtType(c.Type) == ColumnExtType_TIME_CREATE && c.Name != "created_at" {
			log.Errorf("the name of _TIME_CREATED column must be 'created_at'")
			return fmt.Errorf("the name of _TIME_CREATED column must be 'created_at'")
		}
		if ColumnExtType(c.Type) == ColumnExtType_TIME_UPDATE && c.Name != "updated_at" {
			log.Errorf("the name of _TIME_UPDATE column must be 'updated_at'")
			return fmt.Errorf("the name of _TIME_UPDATE column must be 'updated_at'")
		}
	}
	if len(c.ForeignKeys) > 1 {
		log.Errorf("column (%s) foreignKeys count must be 0 or 1", c.Name)
		return fmt.Errorf("column (%s) foreignKeys count must be 0 or 1", c.Name)
	}

	for _, foreignKey := range c.ForeignKeys {
		err := foreignKey.Validate(c)
		if err != nil {
			return err
		}
	}
	cnt := 0
	if c.IsIndex {
		cnt++
	}
	if c.IsUnique {
		cnt++
	}
	if c.IsFullText {
		cnt++

		if !FullTextColumnTypes[c.Type] {
			log.Errorf("column (%s) type (%s) is not available for full text", c.Name, c.Type)
			return fmt.Errorf("column (%s) type (%s) is not available for full text", c.Name, c.Type)
		}
	}
	if cnt > 1 {
		msg := fmt.Sprintf("column (%s) isIndex, isUnique and isFullText can only make one of them to be true at the same time", c.Name)
		log.Error(msg)
		return fmt.Errorf(msg)
	}

	if t.Database.Type != DBType_MYSQL &&
		t.Database.Type != DBType_TIDB {
		if c.Type == ColumnType_TINYBLOB ||
			c.Type == ColumnType_BLOB ||
			c.Type == ColumnType_MEDIUMBLOB ||
			c.Type == ColumnType_LONGBLOB ||
			c.Type == ColumnType_BINARY ||
			c.Type == ColumnType_VARBINARY {
			msg := fmt.Sprintf("column (%s) type (%s) is not available for database type (%s)", c.Name, c.Type, t.Database.Type)
			log.Error(msg)
			return fmt.Errorf(msg)
		}
	} else if t.Database.Type != DBType_PG {
		if c.Type == ColumnType_BOOL ||
			c.Type == ColumnType_TIMETZ ||
			c.Type == ColumnType_TIMESTAMPTZ ||
			c.Type == ColumnType_BYTEA {
			msg := fmt.Sprintf("column (%s) type (%s) is not available for database type (%s)", c.Name, c.Type, t.Database.Type)
			log.Error(msg)
			return fmt.Errorf(msg)
		}
	}

	return nil
}

func (f *ForeignKey) Validate(c *Column) error {
	if f.Table == "" {
		log.Errorf("column (%s) foreign key table is empty", c.Name)
		return fmt.Errorf("column (%s) foreign key table is empty", c.Name)
	}
	return nil
}

func (i *Index) Validate(t *Table) error {
	for _, indexColumn := range i.IndexColumns {
		err := indexColumn.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Unique) Validate(t *Table) error {
	for _, indexColumn := range i.IndexColumns {
		err := indexColumn.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *FullText) Validate(t *Table) error {
	for _, indexColumn := range i.IndexColumns {
		err := indexColumn.Validate()
		if err != nil {
			return err
		}
		for _, col := range t.Columns {
			if col.Name == indexColumn.Name && !FullTextColumnTypes[col.Type] {
				log.Errorf("column (%s) type (%s) is not available for full text", col.Name, col.Type)
				return fmt.Errorf("column (%s) type (%s) is not available for full text", col.Name, col.Type)
			}
		}
	}
	return nil
}

func (ic *IndexColumn) Validate() error {
	if ic.Name == "" {
		log.Errorf("index column name is empty")
		return fmt.Errorf("index column name is empty")
	}
	return nil
}
