package pgoperator

import (
	"fmt"
	"strings"

	cerror "meta-egg/internal/error"

	log "github.com/sirupsen/logrus"
)

func (t *PostgreSQLOperator) GetAllRecords(tableName, pKey string, selects []string) ([]map[string]interface{}, error) {
	if t.DB == nil {
		return nil, cerror.ErrDBNotConnected
	}
	var (
		records []map[string]interface{}
		err     error
	)
	if len(selects) > 0 {
		err = t.DB.Raw(fmt.Sprintf("select \"%s\" from \"%s\" order by \"%s\" asc", strings.Join(selects, "\",\""), tableName, pKey)).Scan(&records).Error
	} else {
		err = t.DB.Raw(fmt.Sprintf("select * from \"%s\" order by \"%s\" asc", tableName, pKey)).Scan(&records).Error
	}
	if err != nil {
		log.Errorf("get all records from table %s failed, err: %v", tableName, err)
		return nil, err
	}
	return records, nil
}
