package commands

import (
	"fmt"

	color "github.com/fatih/color"
	util "github.com/sarumaj/gh-gr/pkg/util"
	supererrors "github.com/sarumaj/go-super/errors"
	cobra "github.com/spf13/cobra"
)

// Version holds the application version.
// It gets filled automatically at build time.
var internalVersion string

// BuildDate holds the date and time at which the application was build.
// It gets filled automatically at build time.
var internalBuildDate string

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Display version information",
	Example: "gh pr version",
	Run: func(*cobra.Command, []string) {
		c := util.Console()

		_ = supererrors.ExceptFn(supererrors.W(fmt.Fprintln(c.Stdout(), c.CheckColors(color.BlueString, "gr version: %s", internalVersion))))
		_ = supererrors.ExceptFn(supererrors.W(fmt.Fprintln(c.Stdout(), c.CheckColors(color.BlueString, "Built at: %s", internalBuildDate))))
	},
}
