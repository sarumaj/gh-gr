package commands

import (
	"fmt"
	"os"
	"time"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	logrus "github.com/sirupsen/logrus"
	cobra "github.com/spf13/cobra"
	pool "gopkg.in/go-playground/pool.v3"
)

var configFlags = &configfile.Configuration{}
var loggerEntry = util.Logger.WithFields(logrus.Fields{"mod": "commands"})

var rootCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gr",
		Short: "gr is a gh cli extension allowing management of multiple repositories at once",
		Run: func(cmd *cobra.Command, args []string) {
			util.FatalIfError(cmd.Help())
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if configfile.ConfigurationExists() {
				configFlags = configfile.Load()
			}

			if configFlags.Verbose {
				util.Logger.SetLevel(logrus.DebugLevel)
			}

			util.Logger.Debug("Running in verbose mode")
		},
		Version: version,
	}

	flags := cmd.PersistentFlags()
	flags.UintVarP(&configFlags.Concurrency, "concurrency", "c", util.GetIdealConcurrency(), "Concurrency for concurrent jobs")
	flags.BoolVarP(&configFlags.Verbose, "verbose", "v", false, "Print verbose logs")
	flags.DurationVarP(&configFlags.Timeout, "timeout", "t", 10*time.Minute, "Set timeout for long running jobs")

	cmd.AddCommand(exportCmd, initCmd, importCmd, pullCmd, pushCmd, removeCmd, statusCmd, updateCmd, versionCmd, viewCmd)

	return cmd
}()

type repositoryOperation func(pool.WorkUnit, *util.Progressbar, *configfile.Configuration, configfile.Repository, *statusList)

func repositoryWorkUnit(fn repositoryOperation, bar *util.Progressbar, conf *configfile.Configuration, repo configfile.Repository, status *statusList) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		fn(wu, bar, conf, repo, status)
		return nil, nil
	}
}

func repositoryOperationLoop(fn repositoryOperation) {
	logger := loggerEntry
	bar := util.NewProgressbar(100)

	exists := configfile.ConfigurationExists()
	logger.Debugf("Config exists: %t", exists)
	if !exists {
		fmt.Fprintln(os.Stderr, util.CheckColors(color.RedString, configfile.ConfigNotFound))
		return
	}

	conf := configfile.Load()
	p := pool.NewLimited(conf.Concurrency)
	defer p.Close()

	batch := p.Batch()

	logger.Debugf("Dispatching %d workers", len(conf.Repositories))

	finished := make(chan bool)
	var status statusList
	go func(finished chan<- bool) {
		for _, repo := range conf.Repositories {
			batch.Queue(repositoryWorkUnit(fn, bar, conf, repo, &status))
		}

		batch.QueueComplete()
		finished <- true
	}(finished)

	go func(finished <-chan bool) {
		for timer := time.NewTimer(conf.Timeout); true; {
			select {

			case <-timer.C:
				batch.Cancel()
				return

			case <-finished:
				return

			}
		}
	}(finished)

	_ = bar.ChangeMax(len(conf.Repositories))
	for range batch.Results() {
		bar.Inc()
	}

	logger.Debug("Collected workers")
	status.print()
}

// Execute executes the root command.
func Execute(Version, BuildDate string) {
	version, buildDate = Version, BuildDate
	util.Logger.Debugf("Version: %s, build date: %s", version, buildDate)

	util.FatalIfError(rootCmd.Execute())
}
