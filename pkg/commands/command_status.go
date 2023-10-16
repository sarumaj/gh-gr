package commands

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
	pool "gopkg.in/go-playground/pool.v3"
)

var statusCmd = func() *cobra.Command {
	var purgeUntracked bool

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show status for all repositories",
		Long: "Show status for all repositories.\n\n" +
			"Per default untracked directories will be only listed.\n" +
			"To clean up untracked directories, provide the \"--cleanup\" option.",
		Example: "gh gr status --cleanup",
		Run: func(*cobra.Command, []string) {
			logger := loggerEntry.WithField("command", "status")

			operationLoop(statusOperation, "Checked")
			conf := configfile.Load()

			logger.Debugf("Purge untracked: %t", purgeUntracked)
			if !purgeUntracked {
				status := newOperationStatus()

				for _, f := range conf.ListUntracked() {
					status.appendErrorRow(f, fmt.Errorf("untracked"))
				}

				status.Sort().Print()
				return
			}

			conf.Cleanup()
		},
	}

	flags := statusCmd.Flags()
	flags.BoolVar(&purgeUntracked, "cleanup", false, "Remove untracked directories")

	return statusCmd
}()

func statusOperation(wu pool.WorkUnit, args operationContext) {
	conf := unwrapOperationContext[*configfile.Configuration](args, "conf")
	repo := unwrapOperationContext[configfile.Repository](args, "repo")
	status := unwrapOperationContext[*operationStatus](args, "status")

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
