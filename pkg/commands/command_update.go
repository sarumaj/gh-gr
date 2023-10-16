package commands

import (
	cobra "github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update configuration",
	Example: "gh pr update",
	Run: func(*cobra.Command, []string) {
		initializeOrUpdateConfig(nil, true)
	},
}
