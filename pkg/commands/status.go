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
		bar := util.NewProgressbar(100).Describe(util.CheckColors(color.BlueString, "Checking..."))
		repositoryOperationLoop(bar, runStatus)
		util.FatalIfError(runLocalStatus())
	},
}

func runStatus(wu pool.WorkUnit, conf *configfile.Configuration, repo configfile.Repository, status *statusList) {
	logger := util.Logger()
	entry := logger.WithField("command", "status")
	if wu.IsCancelled() {
		entry.Warn("work unit has been prematurely canceled")
		return
	}

	var ret string
	if !util.PathExists(repo.Directory) {
		entry.Debugf("Repository %s: path does not exist", repo.Directory)
		status.append(repo.Directory, util.CheckColors(color.RedString, "absent"))
		return
	}

	repository, err := openRepository(repo, status)
	if err != nil {
		entry.Debugf("Repository %s: failed to open: %v", repo.Directory, err)
		return
	}

	head, err := repository.Head()
	if err != nil {
		entry.Debugf("Repository %s: failed to retrieve head: %v", repo.Directory, err)
		status.appendError(repo.Directory, err)
		return
	}

	if branch := head.Name().Short(); branch == repo.Branch {
		ret += util.CheckColors(color.GreenString, branch)
	} else {
		entry.Debugf("Repository %s: unexpected branch", repo.Directory)
		ret += util.CheckColors(color.RedString, branch)
	}

	workTree, err := repository.Worktree()
	if err != nil {
		entry.Debugf("Repository %s: failed to retrieve worktree: %v", repo.Directory, err)
		status.appendError(repo.Directory, err)
		return
	}

	repoStatus, err := workTree.Status()
	if err != nil {
		entry.Debugf("Repository %s: failed to retrieve worktree status: %v", repo.Directory, err)
		status.appendError(repo.Directory, err)
		return
	}

	if repoStatus.IsClean() {
		ret += "\t" + util.CheckColors(color.GreenString, "clean")
	} else {
		entry.Debugf("Repository %s: is dirty", repo.Directory)
		ret += "\t" + util.CheckColors(color.RedString, "dirty")
	}

	remote, err := repository.Remote(git.DefaultRemoteName)
	if err != nil {
		entry.Debugf("Repository %s: failed to retrieve remote name: %v", repo.Directory, err)
		status.appendError(repo.Directory, err)
		return
	}

	remoteRef, err := remote.List(&git.ListOptions{})
	if err != nil {
		entry.Debugf("Repository %s: failed to retrieve remote references: %v", repo.Directory, err)
		status.appendError(repo.Directory, err)
		return
	}

	for _, r := range remoteRef {
		if r.Name().String() == "refs/heads/"+repo.Branch {
			if r.Hash() == head.Hash() {
				entry.Debugf("Repository %s: latest", repo.Directory)
				ret += "\t" + util.CheckColors(color.GreenString, "latest")
			} else {
				entry.Debugf("Repository %s: stale", repo.Directory)
				ret += "\t" + util.CheckColors(color.RedString, "stale")
			}
			break
		}
	}

	status.append(repo.Directory, ret)
}
