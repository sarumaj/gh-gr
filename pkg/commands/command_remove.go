package commands

import (
	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
)

// removeFlags represents the flags for remove command
var removeFlags struct {
	purge bool
}

// removeCmd represents the remove command
var removeCmd = func() *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"reset", "rm", "delete", "del"},
		Short:   "Remove current configuration",
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

			logger.Debugf("Removing config, purge: %t", removeFlags.purge)
			conf.Remove(removeFlags.purge)
		},
	}

	flags := removeCmd.Flags()
	flags.BoolVar(&removeFlags.purge, "purge", false, "DANGER!!! Purge directory with local repositories")

	return removeCmd
}()
