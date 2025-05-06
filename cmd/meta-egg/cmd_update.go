package main

import (
	"github.com/spf13/cobra"
)

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update project",
		RunE:  runUpdate,
	}

	cmd.Flags().StringVarP(&envFlag, FlagEnv, "e", "", "project env file is required e.g: ./env.yml")
	cmd.MarkFlagRequired(FlagEnv)
	cmd.Flags().BoolVar(&uncertainFlag, FlagUncertain, false, "try to replace uncertain files")
	cmd.Flags().StringVar(&templateFlag, FlagTemplate, "", "template files root path")
	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	return updateProject(cmd)
} 