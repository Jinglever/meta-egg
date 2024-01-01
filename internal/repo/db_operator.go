package repo

import (
	"io"

	cerror "meta-egg/internal/error"
	"meta-egg/internal/model"
	"meta-egg/internal/repo/config"
	mysqloperator "meta-egg/internal/repo/mysql_operator"
	pgoperator "meta-egg/internal/repo/postgres_operator"

	log "github.com/sirupsen/logrus"
)

type DBOperator interface {
	/*
	 * 获取数据库配置
	 */
	GetDBConfig() config.DBConfig
	/*
	 * 连接数据库
	 */
	ConnectDB() error
	/*
	 * 关闭数据库连接
	 */
	Close()
	/*
	 * 获取当前数据库的schema
	 */
	GetCurDBSchema() (*model.Database, error)
	/*
	 * 输出更新Schema的SQL
	 * @param targetDBSchema 目标数据库的schema
	 * @param createSQLWriter 输出创建数据库的SQL
	 * @param incSQLWriter 输出更新数据库的增量SQL
	 * @param metaDataSQLWriter 输出数据库的meta表的数据的INSERT SQL
	 */
	OutputSQLForSchemaUpdating(targetDBSchema *model.Database,
		createSQLWriter io.StringWriter,
		incSQLWriter io.StringWriter,
		metaDataSQLWriter io.StringWriter,
	) error
	/*
	 * 获取表的所有记录
	 */
	GetAllRecords(tableName, pKey string, selects []string) ([]map[string]interface{}, error)
}

func NewDBOperator(typ model.DatabaseType, p config.DBConfig, ignoreTables []string) (DBOperator, error) {
	switch typ {
	case model.DBType_MYSQL:
		return mysqloperator.New(p.Host, p.Port, p.User, p.Password, p.DBName, ignoreTables), nil
	case model.DBType_TIDB:
		return mysqloperator.New(p.Host, p.Port, p.User, p.Password, p.DBName, ignoreTables), nil
	case model.DBType_PG:
		return pgoperator.New(p.Host, p.Port, p.User, p.Password, p.DBName, ignoreTables), nil
	default:
		log.Errorf("unsupported database type: %v", typ)
		return nil, cerror.ErrUnsupportedDatabaseType
	}
}
