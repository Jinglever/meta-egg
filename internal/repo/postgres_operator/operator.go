package pgoperator

import (
	"fmt"
	"strings"

	cerror "meta-egg/internal/error"
	"meta-egg/internal/model"
	"meta-egg/internal/repo/config"

	"gorm.io/driver/postgres"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PostgreSQLOperator struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	IgnoreTables map[string]bool // 忽略的表

	DB          *gorm.DB
	CurDBSchema *model.Database
}

func New(dbHost, dbPort, dbUser, dbPassword, dbName string, ignoreTables []string) *PostgreSQLOperator {
	ignoreTableMap := make(map[string]bool)
	for _, t := range ignoreTables {
		ignoreTableMap[t] = true
	}
	return &PostgreSQLOperator{
		DBHost:       dbHost,
		DBPort:       dbPort,
		DBUser:       dbUser,
		DBPassword:   dbPassword,
		DBName:       dbName,
		IgnoreTables: ignoreTableMap,
	}
}

func (t *PostgreSQLOperator) GetDBConfig() config.DBConfig {
	return config.DBConfig{
		Host:     t.DBHost,
		Port:     t.DBPort,
		User:     t.DBUser,
		Password: t.DBPassword,
		DBName:   t.DBName,
	}
}

// Connect to postgresql
func (t *PostgreSQLOperator) ConnectDB() error {
	if t.DB != nil {
		return nil
	}
	// connect to postgresql
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai target_session_attrs=read-write",
		t.DBHost, t.DBPort, t.DBUser, t.DBPassword, t.DBName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "Unknown database") {
			return cerror.ErrDatabaseNotExists
		}

		log.Errorf("connect to postgresql failed, err: %v", err)
		return err
	}
	log.Debugf("database connected")
	t.DB = db
	return nil
}

// Close postgresql connection
func (t *PostgreSQLOperator) Close() {
	if t.DB != nil {
		sqlDB, err := t.DB.DB()
		if err == nil {
			sqlDB.Close()
			log.Debugf("database connection closed")
			t.DB = nil
		}
	}
}
