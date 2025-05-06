package main

import (
	"github.com/spf13/cobra"
)

func newNewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new project",
		RunE:  runNew,
	}

	cmd.Flags().StringVar(&templateFlag, FlagTemplate, "", "template files root path")
	return cmd
}

func runNew(cmd *cobra.Command, args []string) error {
	return newProject(cmd)
} 