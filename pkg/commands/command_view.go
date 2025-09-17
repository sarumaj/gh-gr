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
			"Different output formats supported.\n" +
			"Supports filtering local repositories using glob match:\n\n" +
			"\t- *\t\t\tmatches any sequence of characters besides '/' or '\\' on Windows\n" +
			"\t- ?\t\t\tmatches any single character besides '/' or '\\' on Windows\n" +
			"\t- [ { characters } ]\tcharacter class matching class characters (must be non-empty)\n" +
			"\t- [^ { characters } ]\tcharacter class matching any characters besides class characters (must be non-empty)\n" +
			"\t- c\t\t\tmatches character c (c != '*', '?', '\\', '[')\n" +
			"\t- \\\\c\t\t\tmatches any character c (escaping is disabled on Windows)\n" +
			"\t- [ 'lo' - 'hi' ]\tmatches character c between lo <= c <= hi\n" +
			"\t- [^ 'lo' - 'hi' ]\tmatches any character besides character c between lo <= c <= hi\n",
		Example: "gh pr view -f json",
		Run: func(*cobra.Command, []string) {
			if !configfile.ConfigurationExists() {
				c := util.Console()
				util.PrintlnAndExit("%s", c.CheckColors(color.RedString, configfile.ConfigNotFound))
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
