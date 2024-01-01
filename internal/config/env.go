package config

import (
	jgfile "github.com/Jinglever/go-file"

	jgconf "github.com/Jinglever/go-config"
	"github.com/Jinglever/go-config/option"
	jgstr "github.com/Jinglever/go-string"
	log "github.com/sirupsen/logrus"
)

type EnvConfig struct {
	Project      ProjectConfig  `mapstructure:"project"`
	Manifest     ManifestConfig `mapstructure:"manifest"`
	DB           DBConfig       `mapstructure:"db"`
	IgnoreFiles  []string       `mapstructure:"ignore_files"`  // ignored files will not be cover when update project
	IgnoreTables []string       `mapstructure:"ignore_tables"` // ignored tables will not be diff when make db command
}
type ProjectConfig struct {
	Root string `mapstructure:"root"`
}
type ManifestConfig struct {
	Root string `mapstructure:"root"`
	File string `mapstructure:"file"`
}
type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
}

func LoadEnvFile(envFile string) *EnvConfig {
	// load envFile
	var envCfg EnvConfig
	if err := jgconf.LoadYamlConfig(envFile, &envCfg, option.WithEnvPrefix("EG")); err != nil {
		log.Fatal(err)
	}
	log.Debugf("env config: %s", jgstr.JsonEncode(envCfg))

	if !jgfile.IsDir(envCfg.Project.Root) {
		log.Fatalf("project root dir not exist: %s", envCfg.Project.Root)
	}
	if !jgfile.IsDir(envCfg.Manifest.Root) {
		log.Fatalf("manifest root dir not exist: %s", envCfg.Manifest.Root)
	}
	if !jgfile.IsFile(envCfg.Manifest.File) {
		log.Fatalf("manifest file not exist: %s", envCfg.Manifest.File)
	}
	return &envCfg
}
