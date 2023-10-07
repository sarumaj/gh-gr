package commands

import (
	"fmt"
	"strings"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
)

var viewCmd = func() *cobra.Command {
	var formatOption string

	viewCmd := &cobra.Command{
		Use:   "view",
		Short: "Display current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if !configfile.ConfigurationExists() {
				util.PrintlnAndExit(util.CheckColors(color.RedString, configfile.ConfigNotFound))
			}

			logger := loggerEntry.WithField("command", "view")
			conf := configfile.Load()

			logger.Debug("Streaming")
			conf.Display(formatOption, false)
		},
	}

	flags := viewCmd.Flags()
	supportedFormats := strings.Join(configfile.GetListOfSupportedFormats(true), ", ")
	flags.StringVarP(&formatOption, "format", "f", "yaml", fmt.Sprintf("Change output format, supported formats: [%s]", supportedFormats))

	return viewCmd
}()
