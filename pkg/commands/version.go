package commands

import (
	"fmt"

	semver "github.com/blang/semver"
	color "github.com/fatih/color"
	selfupdate "github.com/rhysd/go-github-selfupdate/selfupdate"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
)

const remoteRepositoryName = "sarumaj/gh-gr"

// Version holds the application version.
// It gets filled automatically at build time.
var version string

// BuildDate holds the date and time at which the application was build.
// It gets filled automatically at build time.
var buildDate string

var versionCmd = func() *cobra.Command {
	var update bool

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			if update {
				selfUpdate()
			} else {
				printVersion()
			}
		},
	}

	flags := versionCmd.Flags()
	flags.BoolVarP(&update, "update", "u", false, "Update extension")

	return versionCmd
}()

func printVersion() {
	current, err := semver.ParseTolerant(version)
	util.FatalIfError(err)

	latest, found, err := selfupdate.DetectLatest(remoteRepositoryName)
	util.FatalIfError(err)

	var vSuffix string
	if !found || latest.Version.LTE(current) {
		vSuffix = "(latest)"
	} else {
		vSuffix = "(newer version available: " + latest.Version.String() + ")"
	}

	fmt.Println(util.CheckColors(color.BlueString, "gr version: %s %s", version, vSuffix))
	fmt.Println(util.CheckColors(color.BlueString, "Built at: %s", buildDate))
}

func selfUpdate() {
	current, err := semver.ParseTolerant(version)
	util.FatalIfError(err)

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Validator: &selfupdate.SHA2Validator{},
	})
	util.FatalIfError(err)

	latest, err := updater.UpdateSelf(current, remoteRepositoryName)
	util.FatalIfError(err)

	if latest.Version.LTE(current) {
		fmt.Println(util.CheckColors(color.BlueString, "You are already using the latest version: %s", version))
	} else {
		fmt.Println(util.CheckColors(color.GreenString, "Successfully updated to version: %s", latest.Version))
	}
}
