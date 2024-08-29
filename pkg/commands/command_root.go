package commands

import (
	"time"

	configfile "github.com/sarumaj/gh-gr/v2/pkg/configfile"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	supererrors "github.com/sarumaj/go-super/errors"
	logrus "github.com/sirupsen/logrus"
	cobra "github.com/spf13/cobra"
)

// configFlags is a global variable holding configuration flags
var configFlags = &configfile.Configuration{}

// loggerEntry is a global variable holding logger entry at package level
var loggerEntry = util.Logger.WithFields(logrus.Fields{"mod": "commands"})

// rootCmd represents the base command when called without any subcommands
var rootCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gr",
		Short: "gr is a gh cli extension allowing management of multiple repositories at once",
		Run: func(cmd *cobra.Command, _ []string) {
			supererrors.Except(cmd.Help())
		},
		Example: "gh gr --concurrency 100 --timeout \"20s\" <subcommand>",
		PersistentPreRun: func(*cobra.Command, []string) {
			if configfile.ConfigurationExists() {
				configFlags = configfile.Load()
			}

			logger := util.Logger
			if util.GetenvBool(util.Verbose) {
				logger.SetLevel(logrus.DebugLevel)
			}

			logger.Debugf("Version: %s, build date: %s, executable path: %s", versionFlags.internalVersion, versionFlags.internalBuildDate, util.GetExecutablePath())
			logger.Debug("Running in verbose mode")
		},
		Version: versionFlags.internalVersion,
	}

	flags := cmd.PersistentFlags()
	flags.UintVarP(&configFlags.Concurrency, "concurrency", "c", util.GetIdealConcurrency(), "Concurrency for concurrent jobs")
	flags.DurationVarP(&configFlags.Timeout, "timeout", "t", 10*time.Minute, "Set timeout for long running jobs")

	cmd.AddCommand(cleanupCmd, editCmd, exportCmd, initCmd, importCmd, pullCmd, pushCmd, prCmd, removeCmd, statusCmd, updateCmd, versionCmd, viewCmd)

	return cmd
}()

// Execute executes the root command.
func Execute(version, buildDate string) {
	versionFlags.internalVersion, versionFlags.internalBuildDate = version, buildDate
	logger := util.Logger

	defer util.AcquireProcessIDLock().Unlock()

	if err := rootCmd.Execute(); err != nil {
		logger.Debugf("Execution failed: %v", err)
	}
}
