package commands

import (
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	cobra "github.com/spf13/cobra"
)

var removeCmd = func() *cobra.Command {
	var purge bool

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			conf := configfile.Load()
			conf.Remove(purge)
		},
	}

	flags := removeCmd.Flags()
	flags.BoolVarP(&purge, "purge", "p", false, "DANGER!!! Purge directory with local repositories")

	return removeCmd
}()
