package commands

import (
	cobra "github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update configuration",
	Run: func(cmd *cobra.Command, args []string) {
		runInit(nil, true)
	},
}
