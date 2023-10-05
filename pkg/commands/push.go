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
	entry := logger.WithField("command", "push")
	if wu.IsCancelled() {
		entry.Warn("work unit has been prematurely canceled")
		return
	}

	repository, err := openRepository(repo, status)
	if err != nil {
		return
	}

	entry.Debugf("Repository %s: pushing to remote", repo.Directory)
	switch err := repository.Push(&git.PushOptions{}); {

	case errors.Is(err, git.ErrNonFastForwardUpdate):
		entry.Debugf("Repository %s: non-fast-forward", repo.Directory)
		status.append(repo.Directory, util.CheckColors(color.RedString, "non-fast-forward update"))
		return

	case
		errors.Is(err, transport.ErrAuthenticationRequired),
		errors.Is(err, transport.ErrAuthorizationFailed):
		entry.Debugf("Repository %s: unauthorized", repo.Directory)

		status.append(repo.Directory, util.CheckColors(color.RedString, "unauthorized"))
		return

	case errors.Is(err, git.NoErrAlreadyUpToDate): // ignore
		entry.Debugf("Repository %s: already up-to-date", repo.Directory)

	case err != nil:
		entry.Debugf("Repository %s: failure: %v", repo.Directory, err)
		status.appendError(repo.Directory, err)
		return

	}

	status.append(repo.Directory, util.CheckColors(color.GreenString, "ok"))
}
