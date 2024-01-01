package main

import (
	"fmt"
	"os"
	"path/filepath"

	"meta-egg/internal/domain/modeler"
	"meta-egg/internal/repo"
	daocfg "meta-egg/internal/repo/config"

	log "github.com/sirupsen/logrus"

	jgfile "github.com/Jinglever/go-file"
	jgstr "github.com/Jinglever/go-string"
	"github.com/urfave/cli/v2"
)

func generateDBSQL(c *cli.Context) error {
	var err error
	checkDebugMode(c)
	cfg := loadEnvConfig(c)

	// parse xml file
	m, err := modeler.ParseXMLFile(cfg.Manifest.File)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("parse xml file failed: %v\n", err)))
		return err
	}

	// create sql dir if not exist
	sqlRoot := filepath.Join(cfg.Manifest.Root, "sql")
	if err = os.MkdirAll(sqlRoot, 0755); err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to create sql dir, err: %v\n", err)))
		return err
	}

	// create db dbOper
	dbName := m.Project.Database.Name
	if cfg.DB.DBName != "" {
		dbName = cfg.DB.DBName
	}
	dbOper, err := repo.NewDBOperator(m.Project.Database.Type,
		daocfg.DBConfig{
			Host:     cfg.DB.Host,
			Port:     cfg.DB.Port,
			User:     cfg.DB.User,
			Password: cfg.DB.Password,
			DBName:   dbName,
		}, cfg.IgnoreTables)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to create db operator, err: %v\n", err)))
		return err
	}
	defer dbOper.Close()

	// try to connect to db
	if dbOper.GetDBConfig().Host != "" {
		err = dbOper.ConnectDB()
		if err == nil {
			log.Debugf("connect db success\n")
			c.App.Writer.Write([]byte(fmt.Sprintf("%s%s%sdatabase exists%s%s, will generate%s %sinc.sql%s\n",
				ColorStatementDiff, FontItalic, FontBold, ColorEnd,
				ColorStatementDiff, ColorEnd,
				ColorRelativeDir, ColorEnd)))
			curDB, err := dbOper.GetCurDBSchema()
			if err != nil {
				c.App.Writer.Write([]byte(fmt.Sprintf("fail to get cur db schema, err: %v\n", err)))
				return err
			}
			log.Debugf("get current db schema success:%s", jgstr.JsonEncode(curDB))
		}
	}

	// use file as io.writer
	f1, _ := os.OpenFile(filepath.Join(sqlRoot, "schema.sql"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	defer f1.Close()
	f2, _ := os.OpenFile(filepath.Join(sqlRoot, "inc.sql"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	defer f2.Close()
	f3, _ := os.OpenFile(filepath.Join(sqlRoot, "meta-data.sql"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	defer f3.Close()
	dbOper.OutputSQLForSchemaUpdating(m.Project.Database, f1, f2, f3)
	c.App.Writer.Write([]byte(fmt.Sprintf("%sgenerate db sql success%s\n", ColorFileDone, ColorEnd)))

	// copy schema.sql and meta-data.sql to <proj_root>/sql/
	c.App.Writer.Write([]byte(fmt.Sprintf("%s%s%scopy%s %sschema.sql%s %s%sand%s %smeta-data.sql%s %s%sto%s %s<proj_root>/sql/%s\n",
		ColorStatementDiff, FontItalic, FontBold, ColorEnd,
		ColorFilesBase, ColorEnd,
		ColorStatementDiff, FontItalic, ColorEnd,
		ColorFilesBase, ColorEnd,
		ColorStatementDiff, FontItalic, ColorEnd,
		ColorRelativeDir, ColorEnd,
	)))
	projRoot := cfg.Project.Root
	projSqlRoot := filepath.Join(projRoot, "sql")
	if err = os.MkdirAll(projSqlRoot, 0755); err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to create sql dir, err: %v\n", err)))
		return err
	}
	if _, err = jgfile.CopyFile(filepath.Join(sqlRoot, "schema.sql"), filepath.Join(projSqlRoot, "schema.sql")); err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to copy schema.sql, err: %v\n", err)))
		return err
	}
	c.App.Writer.Write([]byte(fmt.Sprintf("\t%s%s%s [%s]\n", ColorFileDone, "schema.sql", ColorEnd, GreenCheck)))
	if _, err = jgfile.CopyFile(filepath.Join(sqlRoot, "meta-data.sql"), filepath.Join(projSqlRoot, "meta-data.sql")); err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to copy meta-data.sql, err: %v\n", err)))
		return err
	}
	c.App.Writer.Write([]byte(fmt.Sprintf("\t%s%s%s [%s]\n", ColorFileDone, "meta-data.sql", ColorEnd, GreenCheck)))
	return nil
}
