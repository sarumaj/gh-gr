package commands

import (
	"errors"

	color "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
	"gopkg.in/go-playground/pool.v3"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull all repositories",
	Run: func(cmd *cobra.Command, args []string) {
		bar := util.NewProgressbar(100).Describe(util.CheckColors(color.BlueString, "Pulling..."))
		repositoryOperationLoop(bar, runPull)
	},
}

func runPull(wu pool.WorkUnit, conf *configfile.Configuration, repo configfile.Repository, status *statusList) {
	var repository *git.Repository
	var workTree *git.Worktree
	var err error

	logger := util.Logger()
	if wu.IsCancelled() {
		logger.Warn("work unit has been prematurely canceled")
		return
	}

	conf.Authenticate(&repo.URL)
	conf.Authenticate(&repo.ParentURL)

	if util.PathExists(repo.Directory) {
		repository, err = openRepository(repo, status)
		if err != nil {
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
			status.append(repo.Directory, util.CheckColors(color.RedString, "non-fast-forward update"))
			return

		case errors.Is(err, git.NoErrAlreadyUpToDate): // ignore

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
		if err := pullSubmodule(s); err != nil {
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

	if err := updateRepoConfig(conf, repository); err != nil {
		status.appendError(repo.Directory, err)
		return
	}

	switch _, err := repository.Remote("upstream"); {

	case repo.ParentURL != "" && errors.Is(err, git.ErrRemoteNotFound):
		_, err := repository.CreateRemote(&gitconfig.RemoteConfig{
			Name: "upstream",
			URLs: []string{repo.ParentURL},
		})
		if err != nil {
			status.appendError(repo.Directory, err)
			return
		}

	}

	status.append(repo.Directory, util.CheckColors(color.GreenString, "ok"))
}
