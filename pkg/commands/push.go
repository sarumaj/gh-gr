package commands

import (
	"errors"

	color "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	transport "github.com/go-git/go-git/v5/plumbing/transport"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
	"gopkg.in/go-playground/pool.v3"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push all repositories",
	Run: func(cmd *cobra.Command, args []string) {
		repositoryOperationLoop(runPush)
	},
}

func runPush(wu pool.WorkUnit, bar *util.Progressbar, conf *configfile.Configuration, repo configfile.Repository, status *statusList) {
	interrupt := util.NewInterrupt()
	defer interrupt.Stop()

	logger := loggerEntry.WithField("command", "push").WithField("repository", repo.Directory)

	if bar != nil && conf != nil {
		bar.Describe(util.CheckColors(color.BlueString, conf.GetProgressbarDescriptionForVerb("Pushing", repo)))
	}

	if wu.IsCancelled() {
		logger.Warn("work unit has been prematurely canceled")
		return
	}

	repository, err := openRepository(repo, status)
	if err != nil {
		return
	}

	logger.Debug("Pushing to remote")
	switch err := repository.Push(&git.PushOptions{}); {

	case errors.Is(err, git.ErrNonFastForwardUpdate):
		logger.Debug("Repository is non-fast-forward")
		status.append(repo.Directory, util.CheckColors(color.RedString, "non-fast-forward update"))
		return

	case
		errors.Is(err, transport.ErrAuthenticationRequired),
		errors.Is(err, transport.ErrAuthorizationFailed):
		logger.Debug("Unauthorized")

		status.append(repo.Directory, util.CheckColors(color.RedString, "unauthorized"))
		return

	case errors.Is(err, git.NoErrAlreadyUpToDate): // ignore
		logger.Debug("Repository is already up-to-date")

	case err != nil:
		logger.Debugf("Failure: %v", err)
		status.appendError(repo.Directory, err)
		return

	}

	status.append(repo.Directory, util.CheckColors(color.GreenString, "ok"))
}
