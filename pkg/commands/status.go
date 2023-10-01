package commands

import (
	color "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	configfile "github.com/sarumaj/gh-pr/pkg/configfile"
	util "github.com/sarumaj/gh-pr/pkg/util"
	cobra "github.com/spf13/cobra"
)

var _ = func() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show status for all repositories",
		Run: func(cmd *cobra.Command, args []string) {
			repositoryOperationLoop(runStatus, "Checking")
			runLocalStatus()
		},
	}

	rootCmd.AddCommand(statusCmd)

	return statusCmd
}()

func runStatus(conf *configfile.Configuration, repo configfile.Repository, status *statusList) {
	var ret string

	if !util.PathExists(repo.Directory) {
		status.append(repo.Directory, color.RedString("absent"))

		return
	}

	repository, ok := openRepository(repo, status)
	if !ok {
		return
	}

	head, err := repository.Head()
	if err != nil {
		status.appendError(repo.Directory, err)
		return
	}

	if branch := head.Name().Short(); branch == repo.Branch {
		ret += color.GreenString(branch)
	} else {
		ret += color.RedString(branch)
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
		ret += "\t" + color.GreenString("clean")
	} else {
		ret += "\t" + color.RedString("dirty")
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
				ret += "\t" + color.GreenString("latest")
			} else {
				ret += "\t" + color.RedString("stale")
			}

			break
		}
	}

	status.append(repo.Directory, ret)
}
