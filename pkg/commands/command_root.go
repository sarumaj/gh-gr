package commands

import (
	"time"

	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	logrus "github.com/sirupsen/logrus"
	cobra "github.com/spf13/cobra"
)

var configFlags = &configfile.Configuration{}
var loggerEntry = util.Logger.WithFields(logrus.Fields{"mod": "commands"})

var rootCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gr",
		Short: "gr is a gh cli extension allowing management of multiple repositories at once",
		Run: func(cmd *cobra.Command, _ []string) {
			util.FatalIfError(cmd.Help())
		},
		PersistentPreRun: func(*cobra.Command, []string) {
			if configfile.ConfigurationExists() {
				configFlags = configfile.Load()
			}

			if configFlags.Verbose {
				util.Logger.SetLevel(logrus.DebugLevel)
			}

			util.Logger.Debug("Running in verbose mode")
		},
		Version: internalVersion,
	}

	flags := cmd.PersistentFlags()
	flags.UintVarP(&configFlags.Concurrency, "concurrency", "c", util.GetIdealConcurrency(), "Concurrency for concurrent jobs")
	flags.BoolVarP(&configFlags.Verbose, "verbose", "v", false, "Print verbose logs")
	flags.DurationVarP(&configFlags.Timeout, "timeout", "t", 10*time.Minute, "Set timeout for long running jobs")

	cmd.AddCommand(exportCmd, initCmd, importCmd, pullCmd, pushCmd, removeCmd, statusCmd, updateCmd, versionCmd, viewCmd)

	return cmd
}()

// Execute executes the root command.
func Execute(version, buildDate string) {
	internalVersion, internalBuildDate = version, buildDate
	logger := util.Logger

	logger.Debugf("Version: %s, build date: %s", internalVersion, internalBuildDate)

	if err := rootCmd.Execute(); err != nil {
		logger.Debugf("Execution failed: %v", err)
	}
}
