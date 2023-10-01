package commands

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sarumaj/gh-pr/pkg/configfile"
	"github.com/sarumaj/gh-pr/pkg/util"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/go-playground/pool.v3"
)

var configFlags = &configfile.Configuration{}

var rootCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gr",
		Short: "gr is a gh cli extension allowing management of multiple repositories at once",
		Run: func(cmd *cobra.Command, args []string) {
			util.FatalIfError(cmd.Help())
		},
		Version: Version,
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

	return cmd
}()

type repositoryOperation func(*configfile.Configuration, configfile.Repository, *statusList)

func repositoryWorkUnit(fn repositoryOperation, conf *configfile.Configuration, repo configfile.Repository, status *statusList) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		fn(conf, repo, status)
		return true, nil
	}
}

func repositoryOperationLoop(fn repositoryOperation, msg string) {
	conf := configfile.Load()

	p := pool.NewLimited(conf.Concurrency)
	defer p.Close()

	batch := p.Batch()

	var status statusList
	go func() {
		for _, repo := range conf.Repositories {
			batch.Queue(repositoryWorkUnit(fn, conf, repo, &status))
		}

		batch.QueueComplete()
	}()

	if term.IsTerminal(int(os.Stdout.Fd())) || flag.Lookup("test.v") != nil {
		fmt.Printf("\r%s (0/%d)...", msg, len(conf.Repositories))

		i := 0
		for range batch.Results() {
			fmt.Printf("\r%s (%d/%d)...", msg, i, len(conf.Repositories))
			i++
		}
	}

	fmt.Print(strings.Repeat(" ", len(fmt.Sprintf("\r%[1]s (%[2]d/%[2]d)...", msg, len(conf.Repositories)))) + "\r")

	status.print()
}

// Execute executes the root command.
func Execute() {
	util.FatalIfError(rootCmd.Execute())
}
