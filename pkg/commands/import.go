package commands

import (
	"fmt"
	"strings"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
)

var importCmd = func() *cobra.Command {
	var formatOption string

	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import configuration from stdin",
		Run: func(cmd *cobra.Command, args []string) {
			if configfile.ConfigurationExists() {
				util.PrintlnAndExit(util.CheckColors(color.RedString, configfile.ConfigShouldNotExist))
			}

			logger := loggerEntry.WithField("command", "import")

			logger.Debugf("Import format: %s", formatOption)
			configfile.Import(formatOption)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			updateConfigFlags()
		},
	}

	flags := importCmd.Flags()
	supportedFormats := strings.Join(configfile.GetListOfSupportedFormats(true), ", ")
	flags.StringVarP(&formatOption, "format", "f", "yaml", fmt.Sprintf("Change input format, supported formats: [%s]", supportedFormats))

	return importCmd
}()
