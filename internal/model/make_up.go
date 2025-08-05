package model

import (
	jgstr "github.com/Jinglever/go-string"
)

const (
	CrudUsecaseName = "crud"
)

// 补全冗余信息
func (p *Project) MakeUp() {
	if p.Domain == nil {
		p.Domain = &Domain{}
	}
	p.Domain.MakeUp(p)
	if p.Database != nil {
		p.Database.MakeUp(p)
	}
	if p.ServerType == "" {
		p.ServerType = ServerType_ALL
	}
}

func (d *Domain) MakeUp(p *Project) {
	d.Project = p
	d.UsecaseNameToUsecase = make(map[string]*Usecase, len(d.Usecases))
	for _, usecase := range d.Usecases {
		d.UsecaseNameToUsecase[usecase.Name] = usecase
	}
	for _, usecase := range d.Usecases {
		usecase.MakeUp(d)
	}
}

func (u *Usecase) MakeUp(d *Domain) {
	u.Domain = d
}

func GetDefaultCharset(dbType DatabaseType) string {
	if dbType == DBType_MYSQL || dbType == DBType_TIDB {
		return "utf8mb4"
	} else if dbType == DBType_PG {
		return "UTF8"
	} else {
		return ""
	}
}

func GetDefaultCollate(dbType DatabaseType) string {
	if dbType == DBType_MYSQL || dbType == DBType_TIDB {
		return "utf8mb4_general_ci"
	} else {
		return ""
	}
}

func (d *Database) MakeUp(p *Project) {
	d.Project = p
	d.TableNameToTable = make(map[string]*Table, len(d.Tables))
	if !d.IsSchemaByDB {
		if d.Charset == "" {
			d.Charset = GetDefaultCharset(d.Type)
		}
		if d.Collate == "" {
			d.Collate = GetDefaultCollate(d.Type)
		}
	}
	for _, table := range d.Tables {
		d.TableNameToTable[table.Name] = table
	}
	for _, table := range d.Tables {
		table.MakeUp(d)
	}
}

func (t *Table) MakeUp(d *Database) {
	t.Database = d
	if t.Charset == "" {
		t.Charset = d.Charset // 继承数据库的字符集
	}
	if t.Collate == "" {
		t.Collate = d.Collate // 继承数据库的排序规则
	}
	t.ColNameToColumn = make(map[string]*Column, len(t.Columns))
	for _, column := range t.Columns {
		column.MakeUp(t)
		if column.IsPrimaryKey {
			t.PrimaryColumn = column
		}
		t.ColNameToColumn[column.Name] = column
	}
	var names []string
	// for index
	t.ColNamesToIndex = make(map[string]*Index, len(t.Indexes))
	for _, index := range t.Indexes {
		index.MakeUp(t)
		names = make([]string, 0)
		for _, col := range index.IndexColumns {
			names = append(names, col.Name)
		}
		t.ColNamesToIndex[jgstr.CombineSortedWithSep(",", names)] = index
	}
	// for unique
	t.ColNamesToUnique = make(map[string]*Unique, len(t.Unique))
	for _, unique := range t.Unique {
		unique.MakeUp(t)
		names = make([]string, 0)
		for _, col := range unique.IndexColumns {
			names = append(names, col.Name)
		}
		t.ColNamesToUnique[jgstr.CombineSortedWithSep(",", names)] = unique
	}
	// for fulltext
	t.ColNamesToFullText = make(map[string]*FullText, len(t.FullText))
	for _, fulltext := range t.FullText {
		fulltext.MakeUp(t)
		names = make([]string, 0)
		for _, col := range fulltext.IndexColumns {
			names = append(names, col.Name)
		}
		t.ColNamesToFullText[jgstr.CombineSortedWithSep(",", names)] = fulltext
	}

	// 对于meta表，如果没有指定_DESC字段，那么尝试找一个叫desc的字段
	if t.Type == TableType_META {
		var descCol *Column
		for _, col := range t.Columns {
			if col.ExtType == ColumnExtType_DESC {
				descCol = col
				break
			}
		}
		if descCol == nil {
			for _, col := range t.Columns {
				if col.Name == "desc" && col.Type == ColumnType_VARCHAR {
					col.ExtType = ColumnExtType_DESC
					break
				}
			}
		}
	}

	// 对于RL表和META表，强制设置HasHandler为false，即使XML中配置了handler="true"
	// RL表：通过主表的handler管理，不需要独立的handler
	// META表：静态元数据，不需要复杂的API接口
	if t.Type == TableType_RL || t.Type == TableType_META {
		t.HasHandler = false
	}
}

func (c *Column) MakeUp(t *Table) {
	c.Table = t
	// 转换自定义扩展类型
	if !AvailableColumnTypes[c.Type] {
		if typ, ok := ColumnExtTypeToColumnType[ColumnExtType(c.Type)]; ok {
			c.ExtType = ColumnExtType(c.Type)
			c.Type = typ
		}
	}

	if !c.Table.Database.IsSchemaByDB {
		switch c.ExtType {
		case ColumnExtType_ID:
			c.IsPrimaryKey = true
			c.IsUnsigned = true
		case ColumnExtType_FID:
			c.IsUnsigned = true
		case ColumnExtType_ME_CREATE:
			c.IsUnsigned = true
		case ColumnExtType_ME_UPDATE:
			c.IsUnsigned = true
		case ColumnExtType_ME_DELETE:
			c.IsUnsigned = true
			c.IsRequired = false
			c.IsHidden = true
		case ColumnExtType_TIME_CREATE:
			c.IsRequired = true
		case ColumnExtType_TIME_UPDATE:
			c.IsRequired = true
		case ColumnExtType_TIME_DELETE:
			c.IsRequired = false
			c.IsIndex = true
			c.IsHidden = true
		case ColumnExtType_TIME_DELETE2:
			c.IsRequired = true
			c.IsIndex = true
			c.IsHidden = true
			c.InsertDefault = "0"
			c.IsUnsigned = true
		}

		if c.Table.Database.Type == DBType_PG {
			if c.IsPrimaryKey &&
				c.InsertDefault == "" &&
				IntegerColumnTypes[c.Type] {
				c.InsertDefault = "nextval('" + c.Table.Name + "_" + c.Name + "_seq')"
			}

			if c.Type == ColumnType_TINYINT {
				c.Type = ColumnType_SMALLINT
			}
		}

		if c.Table.Database.Type == DBType_MYSQL ||
			c.Table.Database.Type == DBType_TIDB {
			if c.ExtType == ColumnExtType_BOOL {
				c.Type = ColumnType_TINYINT
			}
			if c.Type == ColumnType_JSONB { // jsonb只有pg支持
				c.Type = ColumnType_JSON
			}
			if c.Type == ColumnType_JSON || c.Type == ColumnType_TEXT {
				if c.InsertDefault != "" {
					c.InsertDefault = "" // mysql不支持json/text类型的默认值
				}
			}
		}

		if c.IsFullText {
			if c.Table.Database.Type != DBType_MYSQL {
				c.IsFullText = false // 暂时只支持mysql的全文索引
			}
		}

		// 新版的mysql貌似也不需要这个了
		//// 补全数值型字段默认显示长度
		// if c.Length == 0 {
		// 	switch c.Type {
		// 	case ColumnType_TINYINT:
		// 		c.Length = 1
		// 	case ColumnType_SMALLINT:
		// 		c.Length = 6
		// 	case ColumnType_MEDIUMINT:
		// 		c.Length = 9
		// 	case ColumnType_INT:
		// 		c.Length = 11
		// 	case ColumnType_BIGINT:
		// 		c.Length = 20
		// 	}
		// }

		if c.Length == 0 {
			switch c.ExtType {
			case ColumnExtType_DESC:
				c.Length = 128
			case ColumnExtType_SMT2:
				c.Length = 16
			}
		}

		if len(c.ForeignKeys) > 0 {
			c.IsIndex = true
		}
	}

	if c.IsPrimaryKey {
		c.IsRequired = true
	}

	if c.Table.Database.Type == DBType_PG {
		if c.Type == ColumnType_BOOL {
			if c.InsertDefault == "0" {
				c.InsertDefault = "false"
			} else if c.InsertDefault == "1" {
				c.InsertDefault = "true"
			}
		}
	}

	for _, foreignKey := range c.ForeignKeys {
		foreignKey.MakeUp(c)
	}
}

func (f *ForeignKey) MakeUp(c *Column) {
	c.IsIndex = true
	f.Column = c
	if c.Table.Database.TableNameToTable[f.Table].PrimaryColumn == nil {
		c.Table.Database.TableNameToTable[f.Table].MakeUp(c.Table.Database)
	}
	if f.Primary == "" {
		f.Primary = c.Table.Database.TableNameToTable[f.Table].PrimaryColumn.Name
	}
}

func (i *Index) MakeUp(t *Table) {
	i.Table = t
}

func (u *Unique) MakeUp(t *Table) {
	u.Table = t
}

func (u *FullText) MakeUp(t *Table) {
	u.Table = t
}
