package commands

import (
	"errors"
	"fmt"

	git "github.com/go-git/go-git/v5"
	transport "github.com/go-git/go-git/v5/plumbing/transport"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
	pool "gopkg.in/go-playground/pool.v3"
)

var pushCmd = &cobra.Command{
	Use:     "push",
	Short:   "Push all repositories",
	Example: "gh pr push",
	Run: func(*cobra.Command, []string) {
		operationLoop(pushOperation, "Pushed")
	},
}

// Push local repository.
func pushOperation(wu pool.WorkUnit, args operationContext) {
	conf := unwrapOperationContext[*configfile.Configuration](args, "conf")
	repo := unwrapOperationContext[configfile.Repository](args, "repo")
	status := unwrapOperationContext[*operationStatus](args, "status")

	logger := loggerEntry.WithField("command", "push").WithField("repository", repo.Directory)

	conf.Authenticate(&repo.URL)
	conf.Authenticate(&repo.ParentURL)
	logger.Debugf("Authenticated: URL: %t, ParentURL: %t", repo.URL != "", repo.ParentURL != "")

	defer util.Chdir(conf.AbsoluteDirectoryPath).Popd()

	logger.Debug("Pushing to remote")
	if err := pushRepository(repo, status); err != nil {
		logger.Debugf("Failed to push: %v", err)
		return
	}

	status.appendStatusRow(repo.Directory, "ok")
}

// Push local repository.
func pushRepository(repo configfile.Repository, status *operationStatus) error {
	repository, err := openRepository(repo, status)
	if err != nil {
		return fmt.Errorf("repository %s: %w", repo.Directory, err)
	}

	switch err := repository.Push(&git.PushOptions{}); {

	case errors.Is(err, git.ErrNonFastForwardUpdate):
		status.appendErrorRow(repo.Directory, fmt.Errorf("non-fast-forward update"))
		return fmt.Errorf("repository %s: %w", repo.Directory, err)

	case
		errors.Is(err, transport.ErrAuthenticationRequired),
		errors.Is(err, transport.ErrAuthorizationFailed):

		status.appendErrorRow(repo.Directory, fmt.Errorf("unauthorized"))
		return fmt.Errorf("repository %s: %w", repo.Directory, err)

	case errors.Is(err, git.NoErrAlreadyUpToDate): // ignore

	case err != nil:
		status.appendErrorRow(repo.Directory, err)
		return fmt.Errorf("repository %s: %w", repo.Directory, err)

	}

	return nil
}
