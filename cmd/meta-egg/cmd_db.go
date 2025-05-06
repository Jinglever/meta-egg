package main

import (
	"github.com/spf13/cobra"
)

func newDBCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Generate db sql",
		RunE:  runDB,
	}

	cmd.Flags().StringVarP(&envFlag, FlagEnv, "e", "", "project env file is required e.g: ./env.yml")
	cmd.MarkFlagRequired(FlagEnv)
	return cmd
}

func runDB(cmd *cobra.Command, args []string) error {
	return generateDBSQL(cmd)
}
