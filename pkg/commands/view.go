package commands

import (
	"fmt"
	"os"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Display current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if !configfile.ConfigurationExists() {
			fmt.Fprintln(os.Stderr, util.CheckColors(color.RedString, configfile.ConfigNotFound))
			return
		}

		logger := loggerEntry.WithField("command", "view")
		conf := configfile.Load()

		logger.Debug("Streaming")
		conf.Display()
	},
}
