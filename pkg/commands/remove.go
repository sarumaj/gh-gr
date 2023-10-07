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
		Run: func(cmd *cobra.Command, args []string) {
			if !configfile.ConfigurationExists() {
				util.PrintlnAndExit(util.CheckColors(color.RedString, configfile.ConfigNotFound))
			}

			interrupt := util.NewInterrupt()
			defer interrupt.Stop()

			logger := loggerEntry.WithField("command", "remove")
			conf := configfile.Load()

			logger.Debugf("Removing config, purge: %t", purge)
			conf.Remove(purge)
		},
	}

	flags := removeCmd.Flags()
	flags.BoolVarP(&purge, "purge", "p", false, "DANGER!!! Purge directory with local repositories")

	return removeCmd
}()
