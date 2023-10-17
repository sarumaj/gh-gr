package commands

import (
	"fmt"

	semver "github.com/blang/semver"
	color "github.com/fatih/color"
	selfupdate "github.com/rhysd/go-github-selfupdate/selfupdate"
	util "github.com/sarumaj/gh-gr/pkg/util"
	supererrors "github.com/sarumaj/go-super/errors"
	cobra "github.com/spf13/cobra"
)

// Address of remote repository where the newest version of gh-gr is released.
const remoteRepository = "sarumaj/gh-gr"

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

		current := supererrors.ExceptFn(supererrors.W(semver.ParseTolerant(internalVersion)))
		latest, found, err := selfupdate.DetectLatest(remoteRepository)

		var vSuffix string
		switch {
		case err == nil && (!found || latest.Version.LTE(current)):
			vSuffix = " (latest)"

		case err == nil && (found || latest.Version.GT(current)):
			vSuffix = " (newer version available: " + latest.Version.String() + ", run \"gh extension upgrade gr\" to update)"

		}

		_ = supererrors.ExceptFn(supererrors.W(fmt.Fprintln(c.Stdout(), c.CheckColors(color.BlueString, "Version: %s", internalVersion+vSuffix))))
		_ = supererrors.ExceptFn(supererrors.W(fmt.Fprintln(c.Stdout(), c.CheckColors(color.BlueString, "Built at: %s", internalBuildDate))))
		_ = supererrors.ExceptFn(supererrors.W(fmt.Fprintln(c.Stdout(), c.CheckColors(color.BlueString, "Executable path: %s", util.GetExecutablePath()))))
	},
}
