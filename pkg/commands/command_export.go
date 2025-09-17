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
var exportFlags struct {
	formatOption string
	output       string
}

// exportCmd represents the export command
var exportCmd = func() *cobra.Command {
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export current configuration to stdout",
		Long: "Export current configuration to stdout.\n\n" +
			"Different output formats supported.",
		Example: " gh gr export --format yaml > export.yaml",
		Run: func(*cobra.Command, []string) {
			if !configfile.ConfigurationExists() {
				c := util.Console()
				util.PrintlnAndExit("%s", c.CheckColors(color.RedString, configfile.ConfigNotFound))
			}

			logger := loggerEntry.WithField("command", "export")
			conf := configfile.Load()

			logger.Debugf("Export format: %s", exportFlags.formatOption)
			conf.Display(exportFlags.formatOption, exportFlags.output, true, nil)
		},
	}

	flags := exportCmd.Flags()
	supportedFormats := strings.Join(configfile.GetListOfSupportedFormats(true), ", ")
	flags.StringVarP(&exportFlags.formatOption, "format", "f", "yaml", fmt.Sprintf("Change output format, supported formats: [%s]", supportedFormats))
	flags.StringVarP(&exportFlags.output, "output", "o", configfile.DefaultExportDestination, "Path to output file or console output (stdout)")

	return exportCmd
}()
