package commands

import (
	"fmt"

	semver "github.com/blang/semver"
	auth "github.com/cli/go-gh/v2/pkg/auth"
	color "github.com/fatih/color"
	selfupdate "github.com/rhysd/go-github-selfupdate/selfupdate"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
)

const (
	remoteRepositoryName = "sarumaj/gh-gr"
	remoteHost           = "github.com"
)

// Version holds the application version.
// It gets filled automatically at build time.
var internalVersion string

// BuildDate holds the date and time at which the application was build.
// It gets filled automatically at build time.
var internalBuildDate string

var versionCmd = func() *cobra.Command {
	var update bool

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(*cobra.Command, []string) {
			logger := loggerEntry.WithField("command", "version")
			logger.Debugf("Update: %t", update)

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

func getUpdater() (updater *selfupdate.Updater, err error) {
	token, _ := auth.TokenForHost(remoteHost)
	if token != "" {
		updater = selfupdate.DefaultUpdater()
		return
	}

	return selfupdate.NewUpdater(selfupdate.Config{
		Validator: &selfupdate.SHA2Validator{},
		APIToken:  token,
	})
}

func printVersion() {
	current, err := semver.ParseTolerant(internalVersion)
	util.FatalIfError(err)

	updater, err := getUpdater()
	util.FatalIfError(err)

	latest, found, err := updater.DetectLatest(remoteRepositoryName)
	util.FatalIfError(err)

	var vSuffix string
	if !found || latest.Version.LTE(current) {
		vSuffix = "(latest)"
	} else {
		vSuffix = "(newer version available: " + latest.Version.String() + ")"
	}

	_, _ = fmt.Fprintln(util.Stdout(), util.CheckColors(color.BlueString, "gr version: %s %s", internalVersion, vSuffix))
	_, _ = fmt.Fprintln(util.Stdout(), util.CheckColors(color.BlueString, "Built at: %s", internalBuildDate))
}

func selfUpdate() {
	defer util.PreventInterrupt()()

	current, err := semver.ParseTolerant(internalVersion)
	util.FatalIfError(err)

	updater, err := getUpdater()
	util.FatalIfError(err)

	latest, err := updater.UpdateSelf(current, remoteRepositoryName)
	util.FatalIfError(err)

	if latest.Version.LTE(current) {
		_, _ = fmt.Fprintln(util.Stdout(), util.CheckColors(color.BlueString, "You are already using the latest version: %s", internalVersion))
	} else {
		_, _ = fmt.Fprintln(util.Stdout(), util.CheckColors(color.GreenString, "Successfully updated to version: %s", latest.Version))
	}
}
