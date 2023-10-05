package commands

import (
	"fmt"
	"os"
	"time"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
	pool "gopkg.in/go-playground/pool.v3"
)

var configFlags = &configfile.Configuration{}

var rootCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gr",
		Short: "gr is a gh cli extension allowing management of multiple repositories at once",
		Run: func(cmd *cobra.Command, args []string) {
			util.FatalIfError(cmd.Help())
		},
		Version: version,
	}

	flags := cmd.PersistentFlags()
	flags.UintVarP(
		&configFlags.Concurrency,
		"concurrency",
		"c",
		util.GetIdealConcurrency(),
		"Concurrency for concurrent jobs",
	)
	flags.BoolVarP(&configFlags.Verbose, "verbose", "v", false, "Print verbose logs")
	flags.DurationVarP(&configFlags.Timeout, "timeout", "t", 10*time.Minute, "Set timeout for long running jobs")

	cmd.AddCommand(initCmd, pullCmd, pushCmd, removeCmd, statusCmd, updateCmd, versionCmd, viewCmd)

	return cmd
}()

type repositoryOperation func(pool.WorkUnit, *configfile.Configuration, configfile.Repository, *statusList)

func repositoryWorkUnit(fn repositoryOperation, conf *configfile.Configuration, repo configfile.Repository, status *statusList) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		fn(wu, conf, repo, status)
		return nil, nil
	}
}

func repositoryOperationLoop(bar *util.Progressbar, fn repositoryOperation) {
	if !configfile.ConfigurationExists() {
		fmt.Fprintln(os.Stderr, util.CheckColors(color.RedString, configfile.ConfigNotFound))
		return
	}

	conf := configfile.Load()
	p := pool.NewLimited(conf.Concurrency)
	defer p.Close()

	batch := p.Batch()

	finished := make(chan bool)

	var status statusList
	go func(finished chan<- bool) {
		for _, repo := range conf.Repositories {
			batch.Queue(repositoryWorkUnit(fn, conf, repo, &status))
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

	status.print()
}

// Execute executes the root command.
func Execute(Version, BuildDate string) {
	version, buildDate = Version, BuildDate
	util.FatalIfError(rootCmd.Execute())
}
