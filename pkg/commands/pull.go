package commands

import (
	"errors"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/sarumaj/gh-pr/pkg/configfile"
	"github.com/sarumaj/gh-pr/pkg/util"
	"github.com/spf13/cobra"
)

var _ = func() *cobra.Command {
	pullCmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull all repositories",
		Run: func(cmd *cobra.Command, args []string) {
			repositoryOperationLoop(runPull, "Pulling")
		},
	}

	rootCmd.AddCommand(pullCmd)

	return pullCmd
}()

func runPull(conf *configfile.Configuration, repo configfile.Repository, status *statusList) {
	var repository *git.Repository
	var workTree *git.Worktree
	var err error

	conf.Authenticate(&repo.URL)
	conf.Authenticate(&repo.ParentURL)

	if util.PathExists(repo.Directory) {
		repository, ok := openRepository(repo, status)
		if !ok {
			return
		}

		workTree, err = repository.Worktree()
		if err != nil {
			status.appendError(repo.Directory, err)
			return
		}

		repoStatus, err := workTree.Status()
		if err != nil {
			status.appendError(repo.Directory, err)
			return
		}

		if !repoStatus.IsClean() {
			status.appendError(repo.Directory, git.ErrWorktreeNotClean)
			return
		}

		switch err = workTree.Pull(&git.PullOptions{
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		}); {

		case errors.Is(err, git.ErrNonFastForwardUpdate):
			status.append(repo.Directory, color.RedString("non-fast-forward update"))
			return

		case errors.Is(err, git.NoErrAlreadyUpToDate):
			// iginore

		case err != nil:
			status.appendError(repo.Directory, err)
			return

		}

	} else {
		repository, err = git.PlainClone(repo.Directory, false, &git.CloneOptions{
			URL:               repo.URL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		if err != nil {
			status.appendError(repo.Directory, err)
			return
		}

		workTree, err = repository.Worktree()
		if err != nil {
			status.appendError(repo.Directory, err)
			return
		}
	}

	submodules, err := workTree.Submodules()
	if err != nil {
		status.appendError(repo.Directory, err)

		return
	}

	for _, s := range submodules {
		err := pullSubmodule(s)
		if err != nil {
			status.appendError(repo.Directory, err)
			return
		}
	}

	err = repository.Fetch(&git.FetchOptions{
		RefSpecs: []gitconfig.RefSpec{"refs/*:refs/*"},
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		status.appendError(repo.Directory, err)
		return
	}

	updateRepoConfig(conf, repository)
	_, err = repository.Remote("upstream")

	if repo.ParentURL != "" && errors.Is(err, git.ErrRemoteNotFound) {
		_, err := repository.CreateRemote(&gitconfig.RemoteConfig{
			Name: "upstream",
			URLs: []string{repo.ParentURL},
		})
		if err != nil {
			status.appendError(repo.Directory, err)
			return
		}
	}

	status.append(repo.Directory, color.GreenString("ok"))
}