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
		bar := util.NewProgressbar(100).Describe(util.CheckColors(color.BlueString, "Pushing..."))
		repositoryOperationLoop(bar, runPush)
	},
}

func runPush(wu pool.WorkUnit, conf *configfile.Configuration, repo configfile.Repository, status *statusList) {
	logger := util.Logger()
	if wu.IsCancelled() {
		logger.Warn("work unit has been prematurely canceled")
		return
	}

	repository, err := openRepository(repo, status)
	if err != nil {
		return
	}

	switch err := repository.Push(&git.PushOptions{}); {

	case errors.Is(err, git.ErrNonFastForwardUpdate):
		status.append(repo.Directory, util.CheckColors(color.RedString, "non-fast-forward update"))
		return

	case
		errors.Is(err, transport.ErrAuthenticationRequired),
		errors.Is(err, transport.ErrAuthorizationFailed):

		status.append(repo.Directory, util.CheckColors(color.RedString, "unauthorized"))
		return

	case errors.Is(err, git.NoErrAlreadyUpToDate): // ignore

	case err != nil:
		status.appendError(repo.Directory, err)
		return

	}

	status.append(repo.Directory, util.CheckColors(color.GreenString, "ok"))
}
