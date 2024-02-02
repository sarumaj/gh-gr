package commands

import (
	cobra "github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update configuration and fetch repositories",
	Example: "gh pr update",
	Run: func(*cobra.Command, []string) {
		initializeOrUpdateConfig(nil, true)
	},
}
