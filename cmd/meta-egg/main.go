package main

import (
	"os"

	"meta-egg/internal/config"
	"meta-egg/pkg/version"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	FlagEnv       = "env"
	FlagRoot      = "root"
	FlagDebug     = "debug"
	FlagUncertain = "uncertain"
	FlagTemplate  = "template"
)

var (
	envFlag = cli.StringFlag{
		Name:     FlagEnv,
		Required: true,
		Aliases:  []string{"e"},
		Usage:    "project env file is required e.g: ./env.yml",
	}
	debugFlag = cli.BoolFlag{
		Name:  FlagDebug,
		Usage: "debug mode",
	}
)

func main() {
	log.SetLevel(log.InfoLevel)
	if version.Release == version.ReleaseIE {
		log.SetReportCaller(true)
	}
	log.SetFormatter(&log.TextFormatter{
		DisableQuote:    true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	app := cli.NewApp()
	app.Name = "meta-egg"
	app.Usage = "meta-egg is a tool to build project from manifest file"
	app.Version = version.GetVersion()
	app.Commands = make([]*cli.Command, 0)

	// new
	newCmd := &cli.Command{
		Name:   "new",
		Usage:  "create a new project",
		Action: newProject,
		Flags: []cli.Flag{
			&debugFlag,
		},
	}
	newCmd.Flags = append(newCmd.Flags, &cli.StringFlag{
		Name:  FlagTemplate,
		Usage: "template files root path",
	})
	app.Commands = append(app.Commands, newCmd)

	// update
	updateCmd := &cli.Command{
		Name:   "update",
		Usage:  "update project",
		Action: updateProject,
		Flags: []cli.Flag{
			&envFlag,
			&debugFlag,
		},
	}
	updateCmd.Flags = append(updateCmd.Flags, &cli.BoolFlag{
		Name:  FlagUncertain,
		Usage: "try to replace uncertain files",
	})
	updateCmd.Flags = append(updateCmd.Flags, &cli.StringFlag{
		Name:  FlagTemplate,
		Usage: "template files root path",
	})
	app.Commands = append(app.Commands, updateCmd)

	// db
	dbCmd := &cli.Command{
		Name:   "db",
		Usage:  "generate db sql",
		Action: generateDBSQL,
		Flags: []cli.Flag{
			&envFlag,
			&debugFlag,
		},
	}
	app.Commands = append(app.Commands, dbCmd)

	// help
	helpCmd := &cli.Command{
		Name:        "help",
		Usage:       "show help info",
		Subcommands: []*cli.Command{},
	}
	helpCmd.Subcommands = append(helpCmd.Subcommands, &cli.Command{
		Name:   "template",
		Usage:  "show template help info",
		Action: showTemplateHelp,
	})
	app.Commands = append(app.Commands, helpCmd)

	_ = app.Run(os.Args)
}

func checkDebugMode(c *cli.Context) {
	if c.Bool(FlagDebug) {
		log.SetLevel(log.DebugLevel)
	}
}

func loadEnvConfig(c *cli.Context) *config.EnvConfig {
	envFile := c.String(FlagEnv)
	return config.LoadEnvFile(envFile)
}
