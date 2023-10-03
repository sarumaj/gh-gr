package commands

import (
	"fmt"
	"os"

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
				fmt.Fprintln(os.Stderr, util.CheckColors(color.RedString, configfile.ConfigNotFound))
				return
			}

			conf := configfile.Load()
			conf.Remove(purge)
		},
	}

	flags := removeCmd.Flags()
	flags.BoolVarP(&purge, "purge", "p", false, "DANGER!!! Purge directory with local repositories")

	return removeCmd
}()
