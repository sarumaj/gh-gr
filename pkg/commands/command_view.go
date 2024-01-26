package commands

import (
	"fmt"
	"strings"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/v2/pkg/configfile"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	cobra "github.com/spf13/cobra"
)

// viewFlags represents the flags for view command
var viewFlags struct {
	formatOption string
	filters      []string
}

// viewCmd represents the view command
var viewCmd = func() *cobra.Command {
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
			conf.Display(viewFlags.formatOption, configfile.DefaultExportDestination, false, viewFlags.filters)
		},
	}

	flags := viewCmd.Flags()
	supportedFormats := strings.Join(configfile.GetListOfSupportedFormats(true), ", ")
	flags.StringVarP(&viewFlags.formatOption, "format", "f", "yaml", fmt.Sprintf("Change output format, supported formats: [%s]", supportedFormats))
	flags.StringArrayVarP(&viewFlags.filters, "match", "m", []string{}, "Glob pattern(s) to filter repositories")

	return viewCmd
}()
