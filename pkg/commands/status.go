package commands

import (
	color "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
	cobra "github.com/spf13/cobra"
	"gopkg.in/go-playground/pool.v3"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status for all repositories",
	Run: func(cmd *cobra.Command, args []string) {
		repositoryOperationLoop(runStatus)
		util.FatalIfError(runLocalStatus())
	},
}

func runStatus(wu pool.WorkUnit, bar *util.Progressbar, conf *configfile.Configuration, repo configfile.Repository, status *statusList) {
	interrupt := util.NewInterrupt()
	defer interrupt.Stop()

	logger := loggerEntry.WithField("command", "status").WithField("repository", repo.Directory)

	if bar != nil && conf != nil {
		bar.Describe(util.CheckColors(color.BlueString, conf.GetProgressbarDescriptionForVerb("Checking", repo)))
	}

	if wu.IsCancelled() {
		logger.Warn("work unit has been prematurely canceled")
		return
	}

	var ret string
	if !util.PathExists(repo.Directory) {
		logger.Debug("Local repository does not exist")
		status.append(repo.Directory, util.CheckColors(color.RedString, "absent"))
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
		status.appendError(repo.Directory, err)
		return
	}

	if branch := head.Name().Short(); branch == repo.Branch {
		ret += util.CheckColors(color.GreenString, branch)
	} else {
		logger.Debugf("Not default branch: %s", branch)
		ret += util.CheckColors(color.RedString, branch)
	}

	workTree, err := repository.Worktree()
	if err != nil {
		logger.Debugf("Failed to retrieve worktree: %v", err)
		status.appendError(repo.Directory, err)
		return
	}

	repoStatus, err := workTree.Status()
	if err != nil {
		logger.Debugf("Failed to retrieve worktree status: %v", err)
		status.appendError(repo.Directory, err)
		return
	}

	if repoStatus.IsClean() {
		ret += "\t" + util.CheckColors(color.GreenString, "clean")
	} else {
		logger.Debug("Repository is dirty")
		ret += "\t" + util.CheckColors(color.RedString, "dirty")
	}

	remote, err := repository.Remote(git.DefaultRemoteName)
	if err != nil {
		logger.Debugf("Failed to retrieve remote name: %v", err)
		status.appendError(repo.Directory, err)
		return
	}

	remoteRef, err := remote.List(&git.ListOptions{})
	if err != nil {
		logger.Debugf("Repository %s: failed to retrieve remote references: %v", repo.Directory, err)
		status.appendError(repo.Directory, err)
		return
	}

	for _, r := range remoteRef {
		if r.Name().String() == "refs/heads/"+repo.Branch {
			if r.Hash() == head.Hash() {
				logger.Debugf("Repository %s: latest", repo.Directory)
				ret += "\t" + util.CheckColors(color.GreenString, "latest")
			} else {
				logger.Debugf("Repository %s: stale", repo.Directory)
				ret += "\t" + util.CheckColors(color.RedString, "stale")
			}
			break
		}
	}

	status.append(repo.Directory, ret)
}
