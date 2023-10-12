package commands

import (
	"errors"
	"fmt"

	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
	pool "gopkg.in/go-playground/pool.v3"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull all repositories",
	Run: func(*cobra.Command, []string) {
		operationLoop(pullOperation)
	},
}

func cloneRemoteRepository(repo configfile.Repository, status *operationStatus) (*git.Repository, *git.Worktree, error) {
	repository, err := git.PlainClone(repo.Directory, false, &git.CloneOptions{
		URL:               repo.URL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		status.appendErrorRow(repo.Directory, err)
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, err)
	}

	workTree, err := repository.Worktree()
	if err != nil {
		status.appendErrorRow(repo.Directory, err)
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, err)
	}

	return repository, workTree, nil
}

func pullExistingRepository(repo configfile.Repository, status *operationStatus) (*git.Repository, *git.Worktree, error) {
	repository, err := openRepository(repo, status)
	if err != nil {
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, err)
	}

	workTree, err := repository.Worktree()
	if err != nil {
		status.appendErrorRow(repo.Directory, err)
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, err)
	}

	repoStatus, err := workTree.Status()
	if err != nil {
		status.appendErrorRow(repo.Directory, err)
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, err)
	}

	if !repoStatus.IsClean() {
		status.appendErrorRow(repo.Directory, git.ErrWorktreeNotClean)
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, git.ErrWorktreeNotClean)
	}

	switch err = workTree.Pull(&git.PullOptions{
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}); {

	case errors.Is(err, git.ErrNonFastForwardUpdate):
		status.appendErrorRow(repo.Directory, fmt.Errorf("non-fast-forward update"))
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, err)

	case errors.Is(err, git.NoErrAlreadyUpToDate): // ignore

	case err != nil:
		status.appendErrorRow(repo.Directory, err)
		return nil, nil, fmt.Errorf("repository %s: %w", repo.Directory, err)

	}

	return repository, workTree, nil
}

func pullOperation(wu pool.WorkUnit, args operationContext) {
	bar := unwrapOperationContext[*util.Progressbar](args, "bar")
	conf := unwrapOperationContext[*configfile.Configuration](args, "conf")
	repo := unwrapOperationContext[configfile.Repository](args, "repo")
	status := unwrapOperationContext[*operationStatus](args, "status")

	defer util.PreventInterrupt()()
	changeProgressbarText(bar, conf, "Pulling", repo)

	logger := loggerEntry.WithField("command", "pull").WithField("repository", repo.Directory)

	conf.Authenticate(&repo.URL)
	conf.Authenticate(&repo.ParentURL)
	logger.Debugf("Authenticated: URL: %t, ParentURL: %t", repo.URL != "", repo.ParentURL != "")

	defer util.MoveToPath(conf.AbsoluteDirectoryPath)()

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

	logger.Debug("Retrieving submodules")
	submodules, err := workTree.Submodules()
	if err != nil {
		logger.Debugf("Failed to retrieve submodules: %v", err)
		status.appendErrorRow(repo.Directory, err)
		return
	}

	logger.Debugf("Pulling %d submodules", len(submodules))
	for _, s := range submodules {
		if err := pullSubmodule(s); err != nil {
			logger.Debugf("Failed to pull submodule: %v", err)
			status.appendErrorRow(repo.Directory, err)
			return
		}
	}

	if err := repository.Fetch(&git.FetchOptions{
		RefSpecs: []gitconfig.RefSpec{"refs/*:refs/*"},
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {

		status.appendErrorRow(repo.Directory, err)
		return
	}

	host := util.GetHostnameFromPath(repo.URL)
	if err := updateRepoConfig(conf, host, repository); err != nil {
		logger.Debugf("Failed to update repo config: %v", err)
		status.appendErrorRow(repo.Directory, err)
		return
	}

	switch _, err := repository.Remote("upstream"); {

	case repo.ParentURL != "" && errors.Is(err, git.ErrRemoteNotFound):
		if _, err := repository.CreateRemote(&gitconfig.RemoteConfig{
			Name: "upstream",
			URLs: []string{repo.ParentURL},
		}); err != nil {

			logger.Debugf("Failed to create mirror: %v", err)
			status.appendErrorRow(repo.Directory, err)
			return
		}

	}

	status.appendStatusRow(repo.Directory, "ok")
}

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

	if err := worktree.Pull(&git.PullOptions{}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {

		// Ignore NoErrAlreadyUpToDate
		return fmt.Errorf("submodule %s: %w", status.Path, err)
	}

	return nil
}

func updateRepoConfig(conf *configfile.Configuration, host string, repository *git.Repository) error {
	repoConf, err := repository.Config()
	if err != nil {
		return err
	}

	section := repoConf.Raw.Section("user")
	profilesMap := conf.Profiles.ToMap()
	profile, ok := profilesMap[host]
	if !ok {
		return fmt.Errorf("no profile for host: %q", host)
	}

	section.SetOption("name", profile.Fullname)
	section.SetOption("email", profile.Email)

	if err := repoConf.Validate(); err != nil {
		return err
	}

	if err := repository.Storer.SetConfig(repoConf); err != nil {
		return err
	}

	return nil
}
