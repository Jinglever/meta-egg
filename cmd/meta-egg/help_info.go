package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func showTemplateHelp(c *cli.Context) error {
	c.App.Writer.Write([]byte(ColorStatementDiff + FontBold + FontItalic + "template placeholder:\n" + ColorEnd))
	placeholders := [][2]string{
		{"%%GO-MODULE%%", "go module name"},
		{"%%PROJECT-NAME%%", "project name"},
		{"%%PROJECT-DESC%%", "project description"},
		{"%%PROJECT-NAME-PKG%%", "project name as package name"},
		{"%%PROJECT-NAME-DIR%%", "project name as directory name"},
		{"%%PROJECT-NAME-STRUCT%%", "project name as struct name"},
	}
	for _, ph := range placeholders {
		c.App.Writer.Write([]byte(fmt.Sprintf("    %s%s%s\t%s%s%s\n",
			ColorStatementNew, ph[0], ColorEnd,
			ColorFilesBase, ph[1], ColorEnd)))
	}
	return nil
}
