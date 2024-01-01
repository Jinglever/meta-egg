package mysqloperator

import (
	"fmt"
	"strings"

	cerror "meta-egg/internal/error"
	"meta-egg/internal/model"
	"meta-egg/internal/repo/config"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQLOperator struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	IgnoreTables map[string]bool // 忽略的表

	DB          *gorm.DB
	CurDBSchema *model.Database
}

func New(dbHost, dbPort, dbUser, dbPassword, dbName string, ignoreTables []string) *MySQLOperator {
	ignoreTableMap := make(map[string]bool)
	for _, t := range ignoreTables {
		ignoreTableMap[t] = true
	}
	return &MySQLOperator{
		DBHost:     dbHost,
		DBPort:     dbPort,
		DBUser:     dbUser,
		DBPassword: dbPassword,
		DBName:     dbName,
	}
}

func (t *MySQLOperator) GetDBConfig() config.DBConfig {
	return config.DBConfig{
		Host:     t.DBHost,
		Port:     t.DBPort,
		User:     t.DBUser,
		Password: t.DBPassword,
		DBName:   t.DBName,
	}
}

// Connect to mysql
func (t *MySQLOperator) ConnectDB() error {
	if t.DB != nil {
		return nil
	}
	// connect to mysql
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s",
		t.DBUser, t.DBPassword, t.DBHost, t.DBPort, t.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "Unknown database") {
			return cerror.ErrDatabaseNotExists
		}

		log.Errorf("connect to mysql failed, err: %v", err)
		return err
	}
	log.Debugf("database connected")
	t.DB = db
	return nil
}

// Close mysql connection
func (t *MySQLOperator) Close() {
	if t.DB != nil {
		sqlDB, err := t.DB.DB()
		if err == nil {
			sqlDB.Close()
			log.Debugf("database connection closed")
			t.DB = nil
		}
	}
}
