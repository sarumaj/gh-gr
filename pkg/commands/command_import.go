package commands

import (
	"fmt"
	"strings"

	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	cobra "github.com/spf13/cobra"
)

var importCmd = func() *cobra.Command {
	var formatOption string

	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import configuration from stdin",
		Run: func(*cobra.Command, []string) {
			logger := loggerEntry.WithField("command", "import")

			logger.Debugf("Import format: %s", formatOption)
			configfile.Import(formatOption)
		},
		PostRun: func(*cobra.Command, []string) {
			updateConfigFlags()
		},
	}

	flags := importCmd.Flags()
	supportedFormats := strings.Join(configfile.GetListOfSupportedFormats(true), ", ")
	flags.StringVarP(&formatOption, "format", "f", "yaml", fmt.Sprintf("Change input format, supported formats: [%s]", supportedFormats))

	return importCmd
}()
