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
