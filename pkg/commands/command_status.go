package commands

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
	configfile "github.com/sarumaj/gh-gr/v2/pkg/configfile"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	cobra "github.com/spf13/cobra"
	pool "gopkg.in/go-playground/pool.v3"
)

// statusFlags represents flags for status command
var statusFlags struct {
	reset bool
}

// statusCmd represents the status command
var statusCmd = func() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show status for all repositories",
		Long: "Show status for all repositories.\n\n" +
			"Additionally, untracked directories will be listed.",
		Example: "gh gr status",
		Run: func(*cobra.Command, []string) {
			operationLoop[configfile.Repository](statusOperation, "Check", operationContextMap{
				"reset":   statusFlags.reset,
				"headers": []string{"Repository", "Branch", "Status", "Remote"},
			})

			conf := configfile.Load()
			status := newOperationStatus()

			status.SetHeader("Repository", "Status")
			for _, f := range conf.ListUntracked() {
				status.appendRow(f, fmt.Errorf("untracked"))
			}

			status.Sort().Print()
		},
	}

	flags := statusCmd.Flags()
	flags.BoolVar(&statusFlags.reset, "reset-all", false, "Perform hard reset against remote for each dirty local repository "+
		"(it will discard all not staged and not committed changes)")

	return statusCmd
}()

// Check status of local repository.
func statusOperation(_ pool.WorkUnit, args operationContext) {
	conf := unwrapOperationContext[*configfile.Configuration](args, "conf")
	repo := unwrapOperationContext[configfile.Repository](args, "object")
	status := unwrapOperationContext[*operationStatus](args, "status")
	reset := unwrapOperationContext[bool](args, "reset")

	logger := loggerEntry.WithField("command", "status").WithField("repository", repo.Directory)

	conf.AuthenticateURL(&repo.URL)
	conf.AuthenticateURL(&repo.ParentURL)
	logger.Debugf("Authenticated: URL: %t, ParentURL: %t", repo.URL != "", repo.ParentURL != "")

	defer util.Chdir(conf.AbsoluteDirectoryPath).Popd()

	var ret []any
	if !util.PathExists(repo.Directory) {
		logger.Debug("Local repository does not exist")
		status.appendRow(repo.Directory, fmt.Errorf("absent"))
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
		status.appendRow(repo.Directory, err)
		return
	}

	if branch := head.Name().Short(); branch == repo.Branch {
		ret = append(ret, branch)
	} else {
		logger.Debugf("Not default branch: %s", branch)
		ret = append(ret, fmt.Errorf("%v", branch))
	}

	workTree, err := repository.Worktree()
	if err != nil {
		logger.Debugf("Failed to retrieve worktree: %v", err)
		status.appendRow(repo.Directory, err)
		return
	}

	repoStatus, err := workTree.Status()
	if err != nil {
		logger.Debugf("Failed to retrieve worktree status: %v", err)
		status.appendRow(repo.Directory, err)
		return
	}

	if repoStatus.IsClean() {
		ret = append(ret, "clean")
	} else if reset {
		if err := resetRepository(workTree, head); err != nil {
			logger.Debugf("Failed to reset repository worktree: %v", err)
			status.appendRow(repo.Directory, err)
			return

		}
		ret = append(ret, "reset")

	} else {
		logger.Debug("Repository is dirty")
		ret = append(ret, fmt.Errorf("dirty"))
	}

	remote, err := repository.Remote(git.DefaultRemoteName)
	if err != nil {
		logger.Debugf("Failed to retrieve remote name: %v", err)
		status.appendRow(repo.Directory, err)
		return
	}

	remoteRef, err := remote.List(&git.ListOptions{})
	if err != nil {
		logger.Debugf("Failed to retrieve remote references: %v", err)
		status.appendRow(repo.Directory, err)
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

	status.appendRow(repo.Directory, ret...)
}
