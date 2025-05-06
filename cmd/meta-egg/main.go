package main

import (
	"os"

	"meta-egg/pkg/version"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	FlagEnv       = "env"
	FlagRoot      = "root"
	FlagDebug     = "debug"
	FlagUncertain = "uncertain"
	FlagTemplate  = "template"
)

var (
	envFlag       string
	debugFlag     bool
	uncertainFlag bool
	templateFlag  string
)

var rootCmd = &cobra.Command{
	Use:     "meta-egg",
	Short:   "meta-egg is a tool to build project from manifest file",
	Version: version.GetVersion(),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set up logging
		log.SetLevel(log.InfoLevel)
		if version.Release == version.ReleaseIE {
			log.SetReportCaller(true)
		}
		log.SetFormatter(&log.TextFormatter{
			DisableQuote:    true,
			TimestampFormat: "2006-01-02 15:04:05",
		})

		// Check debug mode
		if debugFlag {
			log.SetLevel(log.DebugLevel)
		}
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&debugFlag, FlagDebug, false, "debug mode")

	// Add commands
	rootCmd.AddCommand(newNewCommand())
	rootCmd.AddCommand(newUpdateCommand())
	rootCmd.AddCommand(newDBCommand())
	rootCmd.AddCommand(newHelpCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
