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
	if wu.IsCancelled() {
		logger.Warn("work unit has been prematurely canceled")
		return
	}

	var ret string
	if !util.PathExists(repo.Directory) {
		status.append(repo.Directory, util.CheckColors(color.RedString, "absent"))
		return
	}

	repository, err := openRepository(repo, status)
	if err != nil {
		return
	}

	head, err := repository.Head()
	if err != nil {
		status.appendError(repo.Directory, err)
		return
	}

	if branch := head.Name().Short(); branch == repo.Branch {
		ret += util.CheckColors(color.GreenString, branch)
	} else {
		ret += util.CheckColors(color.RedString, branch)
	}

	workTree, err := repository.Worktree()
	if err != nil {
		status.appendError(repo.Directory, err)
		return
	}

	repoStatus, err := workTree.Status()
	if err != nil {
		status.appendError(repo.Directory, err)
		return
	}

	if repoStatus.IsClean() {
		ret += "\t" + util.CheckColors(color.GreenString, "clean")
	} else {
		ret += "\t" + util.CheckColors(color.RedString, "dirty")
	}

	remote, err := repository.Remote(git.DefaultRemoteName)
	if err != nil {
		status.appendError(repo.Directory, err)
		return
	}

	remoteRef, err := remote.List(&git.ListOptions{})
	if err != nil {
		status.appendError(repo.Directory, err)
		return
	}

	for _, r := range remoteRef {
		if r.Name().String() == "refs/heads/"+repo.Branch {
			if r.Hash() == head.Hash() {
				ret += "\t" + util.CheckColors(color.GreenString, "latest")
			} else {
				ret += "\t" + util.CheckColors(color.RedString, "stale")
			}
			break
		}
	}

	status.append(repo.Directory, ret)
}
