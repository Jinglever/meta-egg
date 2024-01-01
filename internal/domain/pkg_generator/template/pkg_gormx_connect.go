package template

import "meta-egg/internal/domain/helper"

var TplPkgGormxConnect string = helper.PH_META_EGG_HEADER + `
package gormx

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Type string ` + "`" + `mapstructure:"type"` + "`" + ` // 数据库类型
	DSN  string ` + "`" + `mapstructure:"dsn"` + "`" + `  // 数据库连接字符串
	User        string         ` + "`" + `mapstructure:"user"` + "`" + `          // 用户名
	Password    string         ` + "`" + `mapstructure:"password"` + "`" + `      // 密码
	Host        string         ` + "`" + `mapstructure:"host"` + "`" + `          // 主机
	Port        string         ` + "`" + `mapstructure:"port"` + "`" + `          // 端口
	Database    string         ` + "`" + `mapstructure:"database"` + "`" + `      // 数据库名
	Timezone    string         ` + "`" + `mapstructure:"timezone"` + "`" + `     // 时区
	LogLevel string ` + "`" + `mapstructure:"log_level"` + "`" + ` // 日志级别 info|warn|error|silent
	MaxOpen      *int           ` + "`" + `mapstructure:"max_open"` + "`" + `       // 最大连接数 <=0表示无限制
	MaxIdle      *int           ` + "`" + `mapstructure:"max_idle"` + "`" + `       // 最大空闲连接数 <=0表示不保留空闲连接 默认值:2
	MaxLifetime  *time.Duration           ` + "`" + `mapstructure:"max_lifetime"` + "`" + `   // 最大连接存活时 <=0代表无限制
	MaxIdleTime  *time.Duration           ` + "`" + `mapstructure:"max_idle_time"` + "`" + `  // 最大空闲时间 <=0代表无限制
}

// ConnectDB 连接数据库
func ConnectDB(cfg *Config) (DB, error) {
	var (
		db  *gorm.DB
		err error
	)
	logLevel := convertLogLevel(cfg.LogLevel)
	gormOpt := &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		SkipDefaultTransaction: true, // 根据官方文档, 可以提高30%性能
		PrepareStmt:            false, // 关闭预编译, 因为gorm的预编译会导致不支持savepoint
		TranslateError:         true,  // 自动翻译错误
	}
	switch cfg.Type {
	case "mysql":
		db, err = connectMySQL(cfg, gormOpt)
	case "postgres":
		db, err = connectPostgreSQL(cfg, gormOpt)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.MaxOpen != nil {
		sqlDB.SetMaxOpenConns(*cfg.MaxOpen)
	}
	if cfg.MaxIdle != nil {
		sqlDB.SetMaxIdleConns(*cfg.MaxIdle)
	}
	if cfg.MaxLifetime != nil {
		sqlDB.SetConnMaxLifetime(*cfg.MaxLifetime)
	}
	if cfg.MaxIdleTime != nil {
		sqlDB.SetConnMaxIdleTime(*cfg.MaxIdleTime)
	}
	return &DBImpl{DB: db}, err
}

// convert log level string to gorm logger level
func convertLogLevel(level string) logger.LogLevel {
	switch level {
	case "info":
		return logger.Info
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	case "silent":
		return logger.Silent
	default:
		return logger.Info
	}
}

func connectMySQL(cfg *Config, opts ...gorm.Option) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&timeout=5s&parseTime=True&loc=%s",
	cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	url.QueryEscape(cfg.Timezone))
	db, err := gorm.Open(mysql.Open(dsn), opts...)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func connectPostgreSQL(cfg *Config, opts ...gorm.Option) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=%s target_session_attrs=read-write",
	cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.Timezone)
	db, err := gorm.Open(postgres.Open(dsn), opts...)
	if err != nil {
		return nil, err
	}
	return db, nil
}
`
