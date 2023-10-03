package commands

import (
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	cobra "github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update configuration",
	Run: func(cmd *cobra.Command, args []string) {
		conf := configfile.Load()
		runInit(conf, true)
	},
}
