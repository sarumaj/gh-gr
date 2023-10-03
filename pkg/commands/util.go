package commands

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	semver "github.com/blang/semver"
	color "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	util "github.com/sarumaj/gh-gr/pkg/util"
)

type status struct {
	Name  string
	State string
}

type statusList []status

func (statuslist *statusList) appendError(repoName string, err error) {
	*statuslist = append(*statuslist, status{
		Name:  repoName,
		State: color.RedString(err.Error()),
	})
}

func (statuslist *statusList) append(repoName, state string) {
	*statuslist = append(*statuslist, status{
		Name:  repoName,
		State: state,
	})
}

func (statuslist *statusList) print() {
	if len(*statuslist) == 0 {
		return
	}

	sort.Slice(*statuslist, func(i, j int) bool {
		return (*statuslist)[i].Name < (*statuslist)[j].Name
	})

	printer := util.TablePrinter()

	for _, s := range *statuslist {
		_ = printer.AddField(s.Name)
		for _, state := range strings.Split(s.State, "\t") {
			_ = printer.AddField(state)
		}
		_ = printer.EndRow()
	}

	_ = printer.AddField(fmt.Sprintf("Total number: %d\n", len(*statuslist))).
		EndRow().
		Render()
}

func currentVersion() semver.Version {
	mmp := versionRegex.ReplaceAllString(Version, "$MMP")
	current, err := semver.Parse(mmp)
	util.FatalIfError(err)

	return current
}

func isRepoDir(path string, repos []configfile.Repository) bool {
	for _, r := range repos {
		if strings.HasPrefix(r.Directory+"/", path+"/") {
			return true
		}
	}

	return false
}

func openRepository(repo configfile.Repository, status *statusList) (*git.Repository, bool) {
	switch repository, err := git.PlainOpen(repo.Directory); {

	// If we get ErrRepositoryNotExists here, it means the repo is broken
	case errors.Is(err, git.ErrRepositoryNotExists):
		status.append(repo.Directory, color.RedString("broken"))
		return nil, false

	case err != nil:
		status.appendError(repo.Directory, err)
		return nil, false

	default:
		return repository, true
	}
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
			if v.Name() == "HEAD" && v.Target() != "" {
				branchRef := v.Target()
				err := repository.Fetch(&git.FetchOptions{
					RefSpecs: []gitconfig.RefSpec{"refs/*:refs/*"},
				})
				if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
					return fmt.Errorf("submodule %s: %w", status.Path, err)
				}

				err = repository.CreateBranch(&gitconfig.Branch{
					Name:   branchRef.Short(),
					Remote: git.DefaultRemoteName,
					Merge:  branchRef,
				})
				if err != nil && !errors.Is(err, git.ErrBranchExists) {
					return fmt.Errorf("submodule %s: %w", status.Path, err)
				}

				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: branchRef,
				})
				if err != nil {
					return fmt.Errorf("submodule %s: %w", status.Path, err)
				}
			}
		}
	}

	err = worktree.Pull(&git.PullOptions{})

	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		// Ignore NoErrAlreadyUpToDate
		return fmt.Errorf("submodule %s: %w", status.Path, err)
	}

	return nil
}

func runLocalStatus() {
	conf := configfile.Load()

	files, err := filepath.Glob(conf.BaseDirectory + "/*")
	util.FatalIfError(err)

	if conf.SubDirectories {
		parents, err := filepath.Glob(conf.BaseDirectory + "/*/*")
		util.FatalIfError(err)
		files = append(files, parents...)
	}

	var status statusList
	for _, f := range files {
		if !isRepoDir(f, conf.Repositories) {
			status.append(f, color.RedString("untracked"))
		}
	}

	status.print()
}

func updateRepoConfig(conf *configfile.Configuration, repository *git.Repository) {
	repoConf, err := repository.Config()
	util.FatalIfError(err)

	section := repoConf.Raw.Section("user")
	section.SetOption("name", conf.Fullname)
	section.SetOption("email", conf.Email)

	util.FatalIfError(repoConf.Validate())
	util.FatalIfError(repository.Storer.SetConfig(repoConf))
}
