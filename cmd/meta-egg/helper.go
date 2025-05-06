package main

import (
	"meta-egg/internal/config"

	"github.com/spf13/cobra"
)

const (
	// ref: https://en.wikipedia.org/wiki/ANSI_escape_code#Colors 8-bit
	ColorEnd           = "\033[0m"
	FontBold           = "\033[1m"
	FontItalic         = "\033[3m"
	ColorRelativeDir   = "\033[38;5;208m"
	ColorStatementDiff = "\033[38;5;180m"
	ColorFilesDiff     = "\033[38;5;186m"
	ColorStatementNew  = "\033[38;5;75m"
	ColorFilesNew      = "\033[38;5;147m"
	ColorStatementBase = "\033[38;5;132m"
	ColorFilesBase     = "\033[38;5;186m"
	ColorFileDone      = "\033[38;5;36m"

	GreenCheck = "\033[32m\u2713\033[0m"
)

func checkDebugMode(cmd *cobra.Command) {
	debugFlag, _ = cmd.Flags().GetBool(FlagDebug)
}

func loadEnvConfig(cmd *cobra.Command) *config.EnvConfig {
	envFile, _ := cmd.Flags().GetString(FlagEnv)
	return config.LoadEnvFile(envFile)
}
