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
	entry := logger.WithField("command", "pull").WithField("repository", repo.Directory)
	if wu.IsCancelled() {
		entry.Warn("work unit has been prematurely canceled")
		return
	}

	entry.Debug("Authenticating")
	conf.Authenticate(&repo.URL)
	conf.Authenticate(&repo.ParentURL)
	entry.Debugf("Authenticated: URL: %t, ParentURL: %t", repo.URL != "", repo.ParentURL != "")

	if util.PathExists(repo.Directory) {
		entry.Debug("Local repository exists")
		repository, err = openRepository(repo, status)
		if err != nil {
			entry.Debugf("Failed to open: %v", err)
			return
		}

		workTree, err = repository.Worktree()
		if err != nil {
			entry.Debugf("Retrieval of worktree failed: %v", err)
			status.appendError(repo.Directory, err)
			return
		}

		repoStatus, err := workTree.Status()
		if err != nil {
			entry.Debugf("Failed to retrieve worktree status: %v", err)
			status.appendError(repo.Directory, err)
			return
		}

		if !repoStatus.IsClean() {
			entry.Debug("Repository is dirty")
			status.appendError(repo.Directory, git.ErrWorktreeNotClean)
			return
		}

		entry.Debug("Pulling repository")
		switch err = workTree.Pull(&git.PullOptions{
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		}); {

		case errors.Is(err, git.ErrNonFastForwardUpdate):
			entry.Debugf("Repository is non-fast-forward")
			status.append(repo.Directory, util.CheckColors(color.RedString, "non-fast-forward update"))
			return

		case errors.Is(err, git.NoErrAlreadyUpToDate): // ignore
			entry.Debugf("Repository is already up-to-date")

		case err != nil:
			entry.Debugf("Failure: %v", err)
			status.appendError(repo.Directory, err)
			return

		}

	} else {
		entry.Debug("Cloning")
		repository, err = git.PlainClone(repo.Directory, false, &git.CloneOptions{
			URL:               repo.URL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		if err != nil {
			entry.Debugf("Cloning failed: %v", err)
			status.appendError(repo.Directory, err)
			return
		}

		workTree, err = repository.Worktree()
		if err != nil {
			entry.Debugf("Retrieval of worktree failed: %v", err)
			status.appendError(repo.Directory, err)
			return
		}
	}

	entry.Debug("Retrieving submodules")
	submodules, err := workTree.Submodules()
	if err != nil {
		entry.Debugf("Failed to retrieve submodules: %v", err)
		status.appendError(repo.Directory, err)
		return
	}
	entry.Debugf("Retrieved %d submodules", len(submodules))

	entry.Debugf("Pulling %d submodules", len(submodules))
	for _, s := range submodules {
		if err := pullSubmodule(s); err != nil {
			status.appendError(repo.Directory, err)
			return
		}
	}

	entry.Debug("Fetching references")
	err = repository.Fetch(&git.FetchOptions{
		RefSpecs: []gitconfig.RefSpec{"refs/*:refs/*"},
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		status.appendError(repo.Directory, err)
		return
	}

	entry.Debug("Updating repository config")
	if err := updateRepoConfig(conf, repository); err != nil {
		status.appendError(repo.Directory, err)
		return
	}

	switch _, err := repository.Remote("upstream"); {

	case repo.ParentURL != "" && errors.Is(err, git.ErrRemoteNotFound):
		entry.Debugf("Creating remote mirror")
		_, err := repository.CreateRemote(&gitconfig.RemoteConfig{
			Name: "upstream",
			URLs: []string{repo.ParentURL},
		})
		if err != nil {
			entry.Debugf("Failed to create mirror: %v", err)
			status.appendError(repo.Directory, err)
			return
		}

	}

	status.append(repo.Directory, util.CheckColors(color.GreenString, "ok"))
}
