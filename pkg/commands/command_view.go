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
	var filters []string

	viewCmd := &cobra.Command{
		Aliases: []string{"show", "list", "ls"},
		Use:     "view",
		Short:   "Display current configuration",
		Long: "Display current configuration.\n\n" +
			"Different output formats supported.",
		Example: "gh pr view -f json",
		Run: func(*cobra.Command, []string) {
			if !configfile.ConfigurationExists() {
				c := util.Console()
				util.PrintlnAndExit(c.CheckColors(color.RedString, configfile.ConfigNotFound))
			}

			logger := loggerEntry.WithField("command", "view")
			conf := configfile.Load()

			logger.Debug("Streaming")
			conf.Display(formatOption, false, filters...)
		},
	}

	flags := viewCmd.Flags()
	supportedFormats := strings.Join(configfile.GetListOfSupportedFormats(true), ", ")
	flags.StringVarP(&formatOption, "format", "f", "yaml", fmt.Sprintf("Change output format, supported formats: [%s]", supportedFormats))
	flags.StringArrayVarP(&filters, "match", "m", []string{}, "Glob pattern(s) to filter repositories")

	return viewCmd
}()
