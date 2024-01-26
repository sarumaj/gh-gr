package commands

import (
	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
)

// cleanupCmd represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:     "cleanup",
	Aliases: []string{"clean", "cl"},
	Short:   "Clean up untracked local repositories",
	Long: "Clean up untracked local repositories.\n\n" +
		"Multiple selection is possible (default: all).",
	Example: "gh gr cleanup",
	Run: func(*cobra.Command, []string) {
		if !configfile.ConfigurationExists() {
			c := util.Console()
			util.PrintlnAndExit(c.CheckColors(color.RedString, configfile.ConfigNotFound))
		}

		conf := configfile.Load()
		conf.Cleanup()
	},
}
