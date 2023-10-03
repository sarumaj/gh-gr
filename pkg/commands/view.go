package commands

import (
	"os"

	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Display current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		conf := configfile.Load()
		util.FatalIfError(yaml.NewEncoder(os.Stdout).Encode(conf))
	},
}
