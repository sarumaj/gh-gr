package commands

import (
	"errors"

	color "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	transport "github.com/go-git/go-git/v5/plumbing/transport"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
	pool "gopkg.in/go-playground/pool.v3"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push all repositories",
	Run: func(*cobra.Command, []string) {
		repositoryOperationLoop(runPush)
	},
}

func runPush(wu pool.WorkUnit, args repositoryOperationArguments) {
	bar := args.bar
	conf := args.conf
	repo := args.repo
	status := args.status

	defer util.PreventInterrupt()()
	changeProgressbarText(bar, conf, "Pushing", repo)

	logger := loggerEntry.WithField("command", "push").WithField("repository", repo.Directory)

	if wu.IsCancelled() {
		logger.Warn("work unit has been prematurely canceled")
		return
	}

	defer util.MoveToPath(conf.AbsoluteDirectoryPath)()

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
