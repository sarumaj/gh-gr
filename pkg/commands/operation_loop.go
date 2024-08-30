package commands

import (
	"strings"
	"time"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/v2/pkg/configfile"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	pool "gopkg.in/go-playground/pool.v3"
)

// Wrapper for repository operations (e.g. pull, push, status).
func operationLoop[U interface {
	configfile.Repository | configfile.PullRequest
}](fn func(pool.WorkUnit, operationContext), verbInfinitive string, args operationContextMap, headers []string, flush bool, source ...U) {
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
	status.SetHeader(headers...)

	worker := func(object U) func(wu pool.WorkUnit) (any, error) {
		return func(wu pool.WorkUnit) (any, error) {
			if wu.IsCancelled() {
				logger.Warn("work unit has been prematurely canceled")
				return nil, wu.Error()
			}

			ctx := operationContextMap{
				"object": object,
				"status": status,
				"conf":   conf,
			}

			for k, v := range args {
				if _, ok := ctx[k]; ok {
					continue
				}
				ctx[k] = v
			}

			fn(wu, newOperationContext(ctx))

			return object, nil
		}
	}

	changeProgressbarText(bar, conf, verbInfinitive+"ing", configfile.Repository{})
	defer util.PreventInterrupt().Stop()

	go func(finished chan<- bool) {
		var target U
		switch any(target).(type) {
		case configfile.Repository:
			for _, object := range conf.Repositories {
				batch.Queue(worker(any(object).(U)))
			}

		default:
			for _, object := range source {
				batch.Queue(worker(object))
			}
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

		object, ok := value.(U)
		if !ok {
			logger.Warnf("expected configfile.Repository got: %T", value)
			continue
		}

		if strings.HasSuffix(verbInfinitive, "e") {
			changeProgressbarText(bar, conf, verbInfinitive+"d", object)
		} else {
			changeProgressbarText(bar, conf, verbInfinitive+"ed", object)
		}
		_ = bar.Inc()
	}

	logger.Debug("Collected workers")
	if flush {
		status.Sort().Align().Print()
	}
}
