package commands

import (
	"errors"
	"fmt"

	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	configfile "github.com/sarumaj/gh-gr/v2/pkg/configfile"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	cobra "github.com/spf13/cobra"
	pool "gopkg.in/go-playground/pool.v3"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:     "pull",
	Short:   "Pull all repositories",
	Example: "gh pr pull",
	Run: func(*cobra.Command, []string) {
		operationLoop[configfile.Repository](pullOperation, "Pull", nil, []string{"Directory", "Status"}, true)
	},
}

// cloneRemoteRepository clones remote repository locally.
func cloneRemoteRepository(repo configfile.Repository, status *operationStatus) (*git.Repository, *git.Worktree, error) {
	repository, err := git.PlainClone(repo.Directory, false, &git.CloneOptions{
		URL:               repo.URL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		status.appendRow(repo.Directory, err)
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, err)
	}

	workTree, err := repository.Worktree()
	if err != nil {
		status.appendRow(repo.Directory, err)
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, err)
	}

	return repository, workTree, nil
}

// pullExistingRepository pulls remote repository.
func pullExistingRepository(repo configfile.Repository, status *operationStatus) (*git.Repository, *git.Worktree, error) {
	repository, err := openRepository(repo, status)
	if err != nil {
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, err)
	}

	workTree, err := repository.Worktree()
	if err != nil {
		status.appendRow(repo.Directory, err)
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, err)
	}

	repoStatus, err := workTree.Status()
	if err != nil {
		status.appendRow(repo.Directory, err)
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, err)
	}

	if !repoStatus.IsClean() {
		status.appendRow(repo.Directory, git.ErrWorktreeNotClean)
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, git.ErrWorktreeNotClean)
	}

	switch err = workTree.Pull(&git.PullOptions{
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}); {

	case errors.Is(err, git.NoErrAlreadyUpToDate): // ignore

	case err != nil:
		status.appendRow(repo.Directory, err)
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, err)

	}

	return repository, workTree, nil
}

// Pull remote repository.
func pullOperation(_ pool.WorkUnit, args operationContext) {
	conf := unwrapOperationContext[*configfile.Configuration](args, "conf")
	repo := unwrapOperationContext[configfile.Repository](args, "object")
	status := unwrapOperationContext[*operationStatus](args, "status")

	logger := loggerEntry.WithField("command", "pull").WithField("repository", repo.Directory)

	conf.AuthenticateURL(&repo.URL)
	conf.AuthenticateURL(&repo.ParentURL)
	logger.Debugf("Authenticated: URL: %t, ParentURL: %t", repo.URL != "", repo.ParentURL != "")

	defer util.Chdir(conf.AbsoluteDirectoryPath).Popd()

	var repository *git.Repository
	var workTree *git.Worktree
	var err error

	if util.PathExists(repo.Directory) {
		logger.Debug("Local repository exists")
		repository, workTree, err = pullExistingRepository(repo, status)

	} else {
		logger.Debug("Cloning")
		repository, workTree, err = cloneRemoteRepository(repo, status)
	}

	if err != nil {
		logger.Debugf("Either pulling or cloning failed: %v", err)
		return
	}

	logger.Debug("Overwriting repo config")
	host := util.GetHostnameFromPath(repo.URL)
	// update remote URL to use current personal access token
	if err := updateRepoConfig(conf, host, repository); err != nil {
		logger.Debugf("Failed to update repo config: %v", err)
		status.appendRow(repo.Directory, err)
		return
	}

	logger.Debug("Retrieving submodules")
	submodules, err := workTree.Submodules()
	if err != nil {
		logger.Debugf("Failed to retrieve submodules: %v", err)
		status.appendRow(repo.Directory, err)
		return
	}

	logger.Debugf("Pulling %d submodules", len(submodules))
	for _, s := range submodules {
		if err := pullSubmodule(s); err != nil {
			logger.Debugf("Failed to pull submodule: %v", err)
			status.appendRow(repo.Directory, err)
			return
		}
	}

	if err := repository.Fetch(&git.FetchOptions{
		RefSpecs: []gitconfig.RefSpec{"refs/*:refs/*"},
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {

		status.appendRow(repo.Directory, err)
		return
	}

	switch _, err := repository.Remote("upstream"); {

	case repo.ParentURL != "" && errors.Is(err, git.ErrRemoteNotFound):
		if _, err := repository.CreateRemote(&gitconfig.RemoteConfig{
			Name: "upstream",
			URLs: []string{repo.ParentURL},
		}); err != nil {

			logger.Debugf("Failed to create mirror: %v", err)
			status.appendRow(repo.Directory, err)
			return
		}

	}

	status.appendRow(repo.Directory, "ok")
}

// Pull GitHub submodule.
func pullSubmodule(submodule *git.Submodule) error {
	status, err := submodule.Status()
	if err != nil {
		return fmt.Errorf("submodule: %w", err)
	}

	repository, err := submodule.Repository()
	if err != nil {
		return fmt.Errorf("submodule %s: %w", status.Path, err)
	}

	worktree, err := repository.Worktree()
	if err != nil {
		return fmt.Errorf("submodule %s: %w", status.Path, err)
	}

	if status.Branch == "" {
		remote, err := repository.Remote(git.DefaultRemoteName)
		if err != nil {
			return fmt.Errorf("submodule %s: %w", status.Path, err)
		}

		remoteRefs, err := remote.List(&git.ListOptions{})
		if err != nil {
			return fmt.Errorf("submodule %s: %w", status.Path, err)
		}

		for _, v := range remoteRefs {
			if v.Name() != "HEAD" || v.Target() == "" {
				continue
			}

			if err := repository.Fetch(&git.FetchOptions{
				RefSpecs: []gitconfig.RefSpec{"refs/*:refs/*"},
			}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {

				return fmt.Errorf("submodule %s: %w", status.Path, err)
			}

			branchRef := v.Target()
			if err := repository.CreateBranch(&gitconfig.Branch{
				Name:   branchRef.Short(),
				Remote: git.DefaultRemoteName,
				Merge:  branchRef,
			}); err != nil && !errors.Is(err, git.ErrBranchExists) {

				return fmt.Errorf("submodule %s: %w", status.Path, err)
			}

			if err := worktree.Checkout(&git.CheckoutOptions{
				Branch: branchRef,
			}); err != nil {

				return fmt.Errorf("submodule %s: %w", status.Path, err)
			}
		}
	}

	switch err := worktree.Pull(&git.PullOptions{}); {

	case err == nil, errors.Is(err, git.NoErrAlreadyUpToDate): // ignore

	default:
		return fmt.Errorf("submodule %s: %w", status.Path, err)
	}

	return nil
}
