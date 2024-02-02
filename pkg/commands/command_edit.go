package commands

import (
	"fmt"
	"strings"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/v2/pkg/configfile"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	cobra "github.com/spf13/cobra"
)

// exportFlags contains flags for import command
var editFlags struct {
	editor       string
	formatOption string
}

// editCmd represents the edit command
var editCmd = func() *cobra.Command {
	editCmd := &cobra.Command{
		Use:     "edit",
		Short:   "Edit configuration",
		Example: "gh gr edit",
		Run: func(*cobra.Command, []string) {
			if !configfile.ConfigurationExists() {
				c := util.Console()
				util.PrintlnAndExit(c.CheckColors(color.RedString, configfile.ConfigNotFound))
			}

			logger := loggerEntry.WithField("command", "export")
			conf := configfile.Load()

			logger.Debugf("Export format: %s", exportFlags.formatOption)
			conf.Edit(editFlags.formatOption, editFlags.editor)
		},
		PostRun: func(*cobra.Command, []string) {
			updateConfigFlags()
		},
	}

	flags := editCmd.Flags()
	supportedFormats := strings.Join(configfile.GetListOfSupportedFormats(true), ", ")
	flags.StringVarP(&editFlags.formatOption, "format", "f", "yaml", fmt.Sprintf("Select config format, supported formats: [%s]", supportedFormats))
	flags.StringVarP(&editFlags.editor, "editor", "e", "vim", "Editor to use")

	return editCmd
}()
