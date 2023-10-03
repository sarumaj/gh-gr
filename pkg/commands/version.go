package commands

import (
	"fmt"

	semver "github.com/blang/semver"
	selfupdate "github.com/rhysd/go-github-selfupdate/selfupdate"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
)

const remoteRepositoryName = "sarumaj/gh-gr"

// Version holds the application version.
// It gets filled automatically at build time.
var Version string

// BuildDate holds the date and time at which the application was build.
// It gets filled automatically at build time.
var BuildDate string

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
	current := semver.MustParse(Version)
	latest, found, err := selfupdate.DetectLatest(remoteRepositoryName)
	util.FatalIfError(err)

	var vSuffix string
	if !found || latest.Version.LTE(current) {
		vSuffix = "(latest)"
	} else {
		vSuffix = "(newer version available: " + latest.Version.String() + ")"
	}

	fmt.Println("gr version:", Version, vSuffix)
	fmt.Println("Built at:", BuildDate)
}

func selfUpdate() {
	current := semver.MustParse(Version)
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Validator: &selfupdate.SHA2Validator{},
	})
	util.FatalIfError(err)

	latest, err := updater.UpdateSelf(current, remoteRepositoryName)
	util.FatalIfError(err)

	if latest.Version.LTE(current) {
		fmt.Println("You are already using the latest version:", Version)
	} else {
		fmt.Println("Successfully updated to version", latest.Version)
	}
}
