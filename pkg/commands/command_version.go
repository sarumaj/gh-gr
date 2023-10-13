package commands

import (
	"fmt"

	color "github.com/fatih/color"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
)

// Version holds the application version.
// It gets filled automatically at build time.
var internalVersion string

// BuildDate holds the date and time at which the application was build.
// It gets filled automatically at build time.
var internalBuildDate string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Run: func(*cobra.Command, []string) {
		c := util.Console()

		_ = util.FatalIfErrorOrReturn(fmt.Fprintln(c.Stdout(), c.CheckColors(color.BlueString, "gr version: %s", internalVersion)))
		_ = util.FatalIfErrorOrReturn(fmt.Fprintln(c.Stdout(), c.CheckColors(color.BlueString, "Built at: %s", internalBuildDate)))
	},
}
