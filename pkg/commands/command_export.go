package commands

import (
	"fmt"
	"strings"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
)

var exportCmd = func() *cobra.Command {
	var formatOption string

	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export current configuration to stdout",
		Run: func(*cobra.Command, []string) {
			if !configfile.ConfigurationExists() {
				c := util.Console()
				util.PrintlnAndExit(c.CheckColors(color.RedString, configfile.ConfigNotFound))
			}

			logger := loggerEntry.WithField("command", "export")
			conf := configfile.Load()

			logger.Debugf("Export format: %s", formatOption)
			conf.Display(formatOption, true)
		},
	}

	flags := exportCmd.Flags()
	supportedFormats := strings.Join(configfile.GetListOfSupportedFormats(true), ", ")
	flags.StringVarP(&formatOption, "format", "f", "yaml", fmt.Sprintf("Change output format, supported formats: [%s]", supportedFormats))

	return exportCmd
}()
