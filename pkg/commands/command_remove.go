package commands

import (
	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
)

var removeCmd = func() *cobra.Command {
	var purge bool

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove current configuration",
		Long: "Remove current configuration.\n\n" +
			"To remove local repositories as well, provide the \"--purge\" option.",
		Example: "gh gr remove --purge",
		Run: func(*cobra.Command, []string) {
			if !configfile.ConfigurationExists() {
				c := util.Console()
				util.PrintlnAndExit(c.CheckColors(color.RedString, configfile.ConfigNotFound))
			}

			logger := loggerEntry.WithField("command", "remove")
			conf := configfile.Load()

			logger.Debugf("Removing config, purge: %t", purge)
			conf.Remove(purge)
		},
	}

	flags := removeCmd.Flags()
	flags.BoolVar(&purge, "purge", false, "DANGER!!! Purge directory with local repositories")

	return removeCmd
}()
