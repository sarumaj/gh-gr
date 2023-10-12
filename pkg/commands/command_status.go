package commands

import (
	"fmt"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
	pool "gopkg.in/go-playground/pool.v3"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status for all repositories",
	Run: func(*cobra.Command, []string) {
		operationLoop(statusOperation)
		util.FatalIfError(runLocalStatus())
	},
}

func runLocalStatus() error {
	conf := configfile.Load()
	util.PathSanitize(&conf.BaseDirectory)

	files, err := filepath.Glob(conf.BaseDirectory + "/*")
	if err != nil {
		return err
	}

	if conf.SubDirectories {
		parents, err := filepath.Glob(conf.BaseDirectory + "/*/*")
		if err != nil {
			return err
		}

		files = append(files, parents...)
	}

	status := newOperationStatus()
	for _, f := range files {
		if !isRepoDir(f, conf.Repositories) {
			status.appendErrorRow(f, fmt.Errorf("untracked"))
		}
	}

	status.Sort().Print()

	return nil
}

func statusOperation(wu pool.WorkUnit, args operationContext) {
	bar := unwrapOperationContext[*util.Progressbar](args, "bar")
	conf := unwrapOperationContext[*configfile.Configuration](args, "conf")
	repo := unwrapOperationContext[configfile.Repository](args, "repo")
	status := unwrapOperationContext[*operationStatus](args, "status")

	defer util.PreventInterrupt().Stop()
	changeProgressbarText(bar, conf, "Checking", repo)

	logger := loggerEntry.WithField("command", "status").WithField("repository", repo.Directory)

	defer util.Chdir(conf.AbsoluteDirectoryPath).Popd()

	var ret []any
	if !util.PathExists(repo.Directory) {
		logger.Debug("Local repository does not exist")
		status.appendErrorRow(repo.Directory, fmt.Errorf("absent"))
		return
	}

	repository, err := openRepository(repo, status)
	if err != nil {
		logger.Debugf("Failed to open: %v", err)
		return
	}

	head, err := repository.Head()
	if err != nil {
		logger.Debugf("Failed to retrieve head: %v", err)
		status.appendErrorRow(repo.Directory, err)
		return
	}

	if branch := head.Name().Short(); branch == repo.Branch {
		ret = append(ret, branch)
	} else {
		logger.Debugf("Not default branch: %s", branch)
		ret = append(ret, fmt.Errorf(branch))
	}

	workTree, err := repository.Worktree()
	if err != nil {
		logger.Debugf("Failed to retrieve worktree: %v", err)
		status.appendErrorRow(repo.Directory, err)
		return
	}

	repoStatus, err := workTree.Status()
	if err != nil {
		logger.Debugf("Failed to retrieve worktree status: %v", err)
		status.appendErrorRow(repo.Directory, err)
		return
	}

	if repoStatus.IsClean() {
		ret = append(ret, "clean")
	} else {
		logger.Debug("Repository is dirty")
		ret = append(ret, fmt.Errorf("dirty"))
	}

	remote, err := repository.Remote(git.DefaultRemoteName)
	if err != nil {
		logger.Debugf("Failed to retrieve remote name: %v", err)
		status.appendErrorRow(repo.Directory, err)
		return
	}

	remoteRef, err := remote.List(&git.ListOptions{})
	if err != nil {
		logger.Debugf("Failed to retrieve remote references: %v", err)
		status.appendErrorRow(repo.Directory, err)
		return
	}

	for _, r := range remoteRef {
		if r.Name().String() == "refs/heads/"+repo.Branch {
			if r.Hash() == head.Hash() {
				logger.Debugf("Repository %s: latest", repo.Directory)
				ret = append(ret, "latest")
			} else {
				logger.Debugf("Repository %s: stale", repo.Directory)
				ret = append(ret, fmt.Errorf("stale"))
			}
			break
		}
	}

	status.appendCustomRow(repo.Directory, ret...)
}
