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
