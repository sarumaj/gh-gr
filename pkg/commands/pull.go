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
		repositoryOperationLoop(runPull)
	},
}

func runPull(wu pool.WorkUnit, bar *util.Progressbar, conf *configfile.Configuration, repo configfile.Repository, status *statusList) {
	interrupt := util.NewInterrupt()
	defer interrupt.Stop()

	logger := loggerEntry.WithField("command", "pull").WithField("repository", repo.Directory)

	if bar != nil && conf != nil {
		bar.Describe(util.CheckColors(color.BlueString, conf.GetProgressbarDescriptionForVerb("Pulling", repo)))
	}

	if wu.IsCancelled() {
		logger.Warn("work unit has been prematurely canceled")
		return
	}

	conf.Authenticate(&repo.URL)
	conf.Authenticate(&repo.ParentURL)
	logger.Debugf("Authenticated: URL: %t, ParentURL: %t", repo.URL != "", repo.ParentURL != "")

	var repository *git.Repository
	var workTree *git.Worktree
	var err error

	if util.PathExists(repo.Directory) {
		logger.Debug("Local repository exists")
		repository, err = openRepository(repo, status)
		if err != nil {
			logger.Debugf("Failed to open: %v", err)
			return
		}

		workTree, err = repository.Worktree()
		if err != nil {
			logger.Debugf("Retrieval of worktree failed: %v", err)
			status.appendError(repo.Directory, err)
			return
		}

		repoStatus, err := workTree.Status()
		if err != nil {
			logger.Debugf("Failed to retrieve worktree status: %v", err)
			status.appendError(repo.Directory, err)
			return
		}

		if !repoStatus.IsClean() {
			logger.Debug("Repository is dirty")
			status.appendError(repo.Directory, git.ErrWorktreeNotClean)
			return
		}

		logger.Debug("Pulling repository")
		switch err = workTree.Pull(&git.PullOptions{
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		}); {

		case errors.Is(err, git.ErrNonFastForwardUpdate):
			logger.Debugf("Repository is non-fast-forward")
			status.append(repo.Directory, util.CheckColors(color.RedString, "non-fast-forward update"))
			return

		case errors.Is(err, git.NoErrAlreadyUpToDate): // ignore
			logger.Debugf("Repository is already up-to-date")

		case err != nil:
			logger.Debugf("Failure: %v", err)
			status.appendError(repo.Directory, err)
			return

		}

	} else {
		logger.Debug("Cloning")
		repository, err = git.PlainClone(repo.Directory, false, &git.CloneOptions{
			URL:               repo.URL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		if err != nil {
			logger.Debugf("Cloning failed: %v", err)
			status.appendError(repo.Directory, err)
			return
		}

		workTree, err = repository.Worktree()
		if err != nil {
			logger.Debugf("Retrieval of worktree failed: %v", err)
			status.appendError(repo.Directory, err)
			return
		}
	}

	logger.Debug("Retrieving submodules")
	submodules, err := workTree.Submodules()
	if err != nil {
		logger.Debugf("Failed to retrieve submodules: %v", err)
		status.appendError(repo.Directory, err)
		return
	}
	logger.Debugf("Retrieved %d submodules", len(submodules))

	logger.Debugf("Pulling %d submodules", len(submodules))
	for _, s := range submodules {
		if err := pullSubmodule(s); err != nil {
			status.appendError(repo.Directory, err)
			return
		}
	}

	logger.Debug("Fetching references")
	err = repository.Fetch(&git.FetchOptions{
		RefSpecs: []gitconfig.RefSpec{"refs/*:refs/*"},
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		status.appendError(repo.Directory, err)
		return
	}

	logger.Debug("Updating repository config")
	host := util.GetHostnameFromPath(repo.URL)
	if err := updateRepoConfig(conf, host, repository); err != nil {
		status.appendError(repo.Directory, err)
		return
	}

	switch _, err := repository.Remote("upstream"); {

	case repo.ParentURL != "" && errors.Is(err, git.ErrRemoteNotFound):
		logger.Debugf("Creating remote mirror")
		_, err := repository.CreateRemote(&gitconfig.RemoteConfig{
			Name: "upstream",
			URLs: []string{repo.ParentURL},
		})
		if err != nil {
			logger.Debugf("Failed to create mirror: %v", err)
			status.appendError(repo.Directory, err)
			return
		}

	}

	status.append(repo.Directory, util.CheckColors(color.GreenString, "ok"))
}
