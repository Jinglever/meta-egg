package main

import (
	"github.com/spf13/cobra"
)

func newHelpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help",
		Short: "Show help info",
	}

	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Show template help info",
		RunE:  runTemplateHelp,
	}
	cmd.AddCommand(templateCmd)

	return cmd
}

func runTemplateHelp(cmd *cobra.Command, args []string) error {
	return showTemplateHelp(cmd)
} 