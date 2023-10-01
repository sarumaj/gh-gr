package commands

import (
	"errors"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/sarumaj/gh-pr/pkg/configfile"
	"github.com/spf13/cobra"
)

var _ = func() *cobra.Command {
	pushCmd := &cobra.Command{
		Use:   "push",
		Short: "Push all repositories",
		Run: func(cmd *cobra.Command, args []string) {
			repositoryOperationLoop(runPush, "Pushing")
		},
	}

	rootCmd.AddCommand(pushCmd)

	return pushCmd
}()

func runPush(conf *configfile.Configuration, repo configfile.Repository, status *statusList) {
	repository, ok := openRepository(repo, status)
	if !ok {
		return
	}

	switch err := repository.Push(&git.PushOptions{}); {

	case errors.Is(err, git.ErrNonFastForwardUpdate):
		status.append(repo.Directory, color.RedString("non-fast-forward update"))
		return

	case
		errors.Is(err, transport.ErrAuthenticationRequired),
		errors.Is(err, transport.ErrAuthorizationFailed):

		status.append(repo.Directory, color.RedString("unauthorized"))
		return

	case errors.Is(err, git.NoErrAlreadyUpToDate):
		// Ignore NoErrAlreadyUpToDate

	case err != nil:
		status.appendError(repo.Directory, err)
		return

	}

	status.append(repo.Directory, color.GreenString("ok"))
}
