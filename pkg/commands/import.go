package commands

import (
	"fmt"
	"os"
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
			logger := loggerEntry.WithField("command", "import")
			if configfile.ConfigurationExists() {
				fmt.Fprintln(os.Stderr, util.CheckColors(color.RedString, configfile.ConfigShouldNotExist))
				return
			}

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
