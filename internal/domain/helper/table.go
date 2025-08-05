package helper

import (
	"meta-egg/internal/model"
	"meta-egg/internal/repo"

	log "github.com/sirupsen/logrus"
)

type ExtTypeCols struct {
	CreatedBy    *model.Column
	UpdatedBy    *model.Column
	DeletedBy    *model.Column
	DeletedAt    *model.Column
	Semantic     *model.Column // 语义化字段 META表特有
	SemanticDesc *model.Column // 语义化描述字段 META表特有
}

func GetExtTypeCols(table *model.Table) *ExtTypeCols {
	var extTypeCols ExtTypeCols
	for _, col := range table.Columns {
		if col.ExtType == model.ColumnExtType_ME_CREATE {
			extTypeCols.CreatedBy = col
		} else if col.ExtType == model.ColumnExtType_ME_UPDATE {
			extTypeCols.UpdatedBy = col
		} else if col.ExtType == model.ColumnExtType_ME_DELETE {
			extTypeCols.DeletedBy = col
		} else if col.ExtType == model.ColumnExtType_TIME_DELETE {
			extTypeCols.DeletedAt = col
		} else if col.ExtType == model.ColumnExtType_TIME_DELETE2 {
			extTypeCols.DeletedAt = col
		} else if col.ExtType == model.ColumnExtType_SMT {
			extTypeCols.Semantic = col
		} else if col.ExtType == model.ColumnExtType_SMT2 {
			extTypeCols.Semantic = col
		} else if col.ExtType == model.ColumnExtType_DESC {
			extTypeCols.SemanticDesc = col
		}
	}
	return &extTypeCols
}

func GetMetaRecords(table *model.Table, eCols *ExtTypeCols, dbOper repo.DBOperator) ([]map[string]interface{}, error) {
	metaRecords := make([]map[string]interface{}, 0)
	if eCols.Semantic != nil && dbOper != nil && dbOper.GetDBConfig().Host != "" {
		err := dbOper.ConnectDB()
		if err == nil {
			// get meta data
			metaRecords, err = dbOper.GetAllRecords(table.Name, table.PrimaryColumn.Name, []string{})
			if err != nil {
				log.Errorf("get meta data failed: %v", err)
				return nil, err
			}
		}
	}
	return metaRecords, nil
}

// IdentifyRLMainTable 通过外键自动识别RL表的主表
func IdentifyRLMainTable(rlTable *model.Table, tableNameToTable map[string]*model.Table) *model.Table {
	var dataTypeKeys []*model.ForeignKey
	var mainMarkedKey *model.ForeignKey

	// 收集所有外键信息
	for _, column := range rlTable.Columns {
		for _, fk := range column.ForeignKeys {
			if fk.IsMain {
				if mainMarkedKey != nil {
					// 错误：多个主外键标记
					return nil
				}
				mainMarkedKey = fk
			}
			if mainTable := tableNameToTable[fk.Table]; mainTable != nil && mainTable.Type == model.TableType_DATA {
				dataTypeKeys = append(dataTypeKeys, fk)
			}
		}
	}

	// 如果有明确的主外键标记，使用它
	if mainMarkedKey != nil {
		if mainTable := tableNameToTable[mainMarkedKey.Table]; mainTable != nil && mainTable.Type == model.TableType_DATA {
			return mainTable
		}
		return nil
	}

	// 如果只有一个DATA类型外键，自动识别
	if len(dataTypeKeys) == 1 {
		return tableNameToTable[dataTypeKeys[0].Table]
	}

	// 如果有多个DATA类型外键但没有明确标记，报错
	if len(dataTypeKeys) > 1 {
		return nil
	}

	// 没有任何DATA类型外键
	return nil
}

// ShouldIncludeRLInList 判断RL表是否应该在主表列表查询中包含
// 只有当RL表中有字段list=true时，才在主表list方法中自动拉取RL数据
func ShouldIncludeRLInList(rlTable *model.Table) bool {
	for _, column := range rlTable.Columns {
		if column.IsList {
			return true
		}
	}
	return false
}

// GetRLListColumns 获取RL表中设置了list=true的字段
func GetRLListColumns(rlTable *model.Table) []*model.Column {
	listColumns := make([]*model.Column, 0)
	for _, column := range rlTable.Columns {
		if column.IsList {
			listColumns = append(listColumns, column)
		}
	}
	return listColumns
}

// GetMainTableRLs 获取主表的所有RL表
func GetMainTableRLs(mainTable *model.Table, allTables []*model.Table) []*model.Table {
	rlTables := make([]*model.Table, 0)
	for _, table := range allTables {
		if table.Type == model.TableType_RL {
			// 检查这个RL表是否依附于当前主表
			isRelatedToMainTable := false
			for _, column := range table.Columns {
				for _, foreignKey := range column.ForeignKeys {
					if foreignKey.Table == mainTable.Name {
						isRelatedToMainTable = true
						break
					}
				}
				if isRelatedToMainTable {
					break
				}
			}
			if isRelatedToMainTable {
				rlTables = append(rlTables, table)
			}
		}
	}
	return rlTables
}

// ================= BR Table Helper Functions =================

// GetMainTableBRs 获取与主表相关的所有BR表
// BR表是多对多关系表，一个主表可能通过多个BR表与其他DATA表建立关系
func GetMainTableBRs(mainTable *model.Table, allTables []*model.Table) []*model.Table {
	brTables := make([]*model.Table, 0)
	for _, table := range allTables {
		if table.Type == model.TableType_BR {
			// 检查这个BR表是否包含指向当前主表的外键
			isRelatedToMainTable := false
			for _, column := range table.Columns {
				for _, foreignKey := range column.ForeignKeys {
					if foreignKey.Table == mainTable.Name {
						isRelatedToMainTable = true
						break
					}
				}
				if isRelatedToMainTable {
					break
				}
			}
			if isRelatedToMainTable {
				brTables = append(brTables, table)
			}
		}
	}
	return brTables
}

// BRRelatedTables 表示BR表连接的两个DATA表的信息
type BRRelatedTables struct {
	Table1   *model.Table  // 第一个相关的DATA表
	Table2   *model.Table  // 第二个相关的DATA表
	Table1FK *model.Column // 指向Table1的外键列
	Table2FK *model.Column // 指向Table2的外键列
	IsValid  bool          // 是否是有效的BR关系（恰好有2个指向DATA表的外键）
}

// IdentifyBRRelatedTables 识别BR表连接的两个DATA表
// BR表应该恰好有两个外键指向两个不同的DATA表
func IdentifyBRRelatedTables(brTable *model.Table, tableNameToTable map[string]*model.Table) *BRRelatedTables {
	result := &BRRelatedTables{
		IsValid: false,
	}

	var dataForeignKeys []*model.Column

	// 收集所有指向DATA表的外键
	for _, column := range brTable.Columns {
		for _, fk := range column.ForeignKeys {
			if targetTable := tableNameToTable[fk.Table]; targetTable != nil && targetTable.Type == model.TableType_DATA {
				dataForeignKeys = append(dataForeignKeys, column)
				break // 一个列只能有一个外键，找到后跳出
			}
		}
	}

	// BR表必须恰好有两个指向DATA表的外键
	if len(dataForeignKeys) != 2 {
		log.Debugf("BR table (%s) has %d foreign keys to DATA tables, expected exactly 2", brTable.Name, len(dataForeignKeys))
		return result
	}

	// 获取两个外键指向的表
	fk1 := dataForeignKeys[0].ForeignKeys[0] // 第一个外键
	fk2 := dataForeignKeys[1].ForeignKeys[0] // 第二个外键

	table1 := tableNameToTable[fk1.Table]
	table2 := tableNameToTable[fk2.Table]

	// 确保两个外键指向不同的表
	if table1.Name == table2.Name {
		log.Debugf("BR table (%s) has two foreign keys pointing to the same table (%s)", brTable.Name, table1.Name)
		return result
	}

	// 设置结果
	result.Table1 = table1
	result.Table2 = table2
	result.Table1FK = dataForeignKeys[0]
	result.Table2FK = dataForeignKeys[1]
	result.IsValid = true

	return result
}

// GetBRForeignKeyForTable 获取BR表中指向特定目标表的外键列
// 返回外键列的信息，如果没有找到则返回nil
func GetBRForeignKeyForTable(brTable *model.Table, targetTable *model.Table) *model.Column {
	for _, column := range brTable.Columns {
		for _, fk := range column.ForeignKeys {
			if fk.Table == targetTable.Name {
				return column
			}
		}
	}
	return nil
}

// GetBROtherTable 给定BR表和其中一个相关表，获取另一个相关表
// 这个函数用于在已知一边的情况下找到关系的另一边
func GetBROtherTable(brTable *model.Table, knownTable *model.Table, tableNameToTable map[string]*model.Table) *model.Table {
	relatedTables := IdentifyBRRelatedTables(brTable, tableNameToTable)
	if !relatedTables.IsValid {
		return nil
	}

	if relatedTables.Table1.Name == knownTable.Name {
		return relatedTables.Table2
	} else if relatedTables.Table2.Name == knownTable.Name {
		return relatedTables.Table1
	}

	return nil
}

// ShouldIncludeBRRelatedTableInList 判断通过BR关系查询的关联表是否应该包含list字段
// BR表不会出现在主表列表中，而是通过GetRelated{OtherTable}List()独立查询
// 这个函数检查关联表（另一边的DATA表）中是否有list=true的字段
func ShouldIncludeBRRelatedTableInList(relatedTable *model.Table) bool {
	for _, column := range relatedTable.Columns {
		if column.IsList {
			return true
		}
	}
	return false
}

// GetBRRelatedTableListColumns 获取通过BR关系查询的关联表中设置了list=true的字段
// 这些字段将在GetRelated{OtherTable}List()查询中返回
func GetBRRelatedTableListColumns(relatedTable *model.Table) []*model.Column {
	listColumns := make([]*model.Column, 0)
	for _, column := range relatedTable.Columns {
		if column.IsList {
			listColumns = append(listColumns, column)
		}
	}
	return listColumns
}
