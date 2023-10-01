package commands

import (
	"github.com/sarumaj/gh-pr/pkg/configfile"
	"github.com/spf13/cobra"
)

var _ = func() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update configuration",
		Run: func(cmd *cobra.Command, args []string) {
			conf := configfile.Load()
			runInit(conf, true)
		},
	}

	rootCmd.AddCommand(updateCmd)

	return updateCmd
}()
