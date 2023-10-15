package commands

import (
	"time"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	pool "gopkg.in/go-playground/pool.v3"
)

func operationLoop(fn func(pool.WorkUnit, operationContext), verb string) {
	logger := loggerEntry
	bar := util.NewProgressbar(100)

	exists := configfile.ConfigurationExists()
	logger.Debugf("Config exists: %t", exists)
	if !exists {
		c := util.Console()
		util.PrintlnAndExit(c.CheckColors(color.RedString, configfile.ConfigNotFound))
	}

	conf := configfile.Load()
	p := pool.NewLimited(conf.Concurrency)
	defer p.Close()

	batch := p.Batch()

	logger.Debugf("Dispatching %d workers", len(conf.Repositories))

	finished := make(chan bool)
	status := newOperationStatus()

	worker := func(repo configfile.Repository) func(wu pool.WorkUnit) (any, error) {
		return func(wu pool.WorkUnit) (any, error) {
			if wu.IsCancelled() {
				logger.Warn("work unit has been prematurely canceled")
				return nil, wu.Error()
			}

			fn(wu, newOperationContext(operationContextMap{
				"conf":   conf,
				"repo":   repo,
				"status": status,
			}))

			return repo, nil
		}
	}

	defer util.PreventInterrupt().Stop()

	go func(finished chan<- bool) {
		for _, repo := range conf.Repositories {
			batch.Queue(worker(repo))
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
	for result := range batch.Results() {
		value, err := result.Value(), result.Error()
		if err != nil {
			logger.Warnf("worker returned error: %v", err)
			continue
		}

		repo, ok := value.(configfile.Repository)
		if !ok {
			logger.Warnf("expected configfile.Repository got: %T", value)
			continue
		}

		changeProgressbarText(bar, conf, verb, repo)
		_ = bar.Inc()
	}

	logger.Debug("Collected workers")
	status.Sort().Print()
}
