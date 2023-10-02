package commands

import (
	"github.com/sarumaj/gh-gr/pkg/configfile"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update configuration",
	Run: func(cmd *cobra.Command, args []string) {
		conf := configfile.Load()
		runInit(conf, true)
	},
}
