package commands

import (
	"fmt"
	"strings"

	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	cobra "github.com/spf13/cobra"
)

// importFlags contains flags for import command
var importFlags struct {
	formatOption string
}

// importCmd represents the import command
var importCmd = func() *cobra.Command {
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import configuration from stdin or a file",
		Long: "Import configuration from stdin or a file.\n\n" +
			"Different output formats supported.\n" +
			"Command supports piped input and HEREDOC.\n" +
			"Caution! The configuration will be overwritten!",
		Example: "cat export.yaml | gh gr import --format yaml",
		Run: func(*cobra.Command, []string) {
			logger := loggerEntry.WithField("command", "import")
			logger.Debugf("Import format: %s", importFlags.formatOption)

			configfile.Import(importFlags.formatOption)
		},
		PostRun: func(*cobra.Command, []string) {
			updateConfigFlags()
		},
	}

	flags := importCmd.Flags()
	supportedFormats := strings.Join(configfile.GetListOfSupportedFormats(true), ", ")
	flags.StringVarP(&importFlags.formatOption, "format", "f", "yaml", fmt.Sprintf("Change input format, supported formats: [%s]", supportedFormats))

	return importCmd
}()
