package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	auth "github.com/cli/go-gh/v2/pkg/auth"
	color "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	selfupdate "github.com/rhysd/go-github-selfupdate/selfupdate"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	extras "github.com/sarumaj/gh-gr/pkg/extras"
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
		State: util.CheckColors(color.RedString, err.Error()),
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

	util.FatalIfError(printer.AddField(fmt.Sprintf("Total number: %d\n", len(*statuslist))).
		EndRow().
		Render())
}

func addGitAliases() error {
	var ga []struct {
		Alias   string `json:"alias"`
		Command string `json:"command"`
	}
	if err := json.Unmarshal(extras.AliasesJSON, &ga); err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	gitconfigPath := filepath.Join(home, ".gitconfig")
	gitconfigRaw, err := os.ReadFile(gitconfigPath)
	if err != nil {
		return err
	}

	cfg := gitconfig.NewConfig()
	if err := cfg.Unmarshal(gitconfigRaw); err != nil {
		return err
	}

	section := cfg.Raw.Section("alias")
	for _, alias := range ga {
		section.SetOption(alias.Alias, alias.Command)
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	gitconfigNew, err := cfg.Marshal()
	if err != nil {
		return err
	}

	if err := os.WriteFile(gitconfigPath, gitconfigNew, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func getUpdater() (updater *selfupdate.Updater, err error) {
	token, _ := auth.TokenForHost(remoteHost)
	if token != "" {
		updater = selfupdate.DefaultUpdater()
		return
	}

	return selfupdate.NewUpdater(selfupdate.Config{
		Validator: &selfupdate.SHA2Validator{},
		APIToken:  token,
	})
}

func isRepoDir(path string, repos []configfile.Repository) bool {
	util.PathSanitize(&path)
	for _, r := range repos {
		util.PathSanitize(&r.Directory)
		if strings.HasPrefix(r.Directory+"/", path+"/") {
			return true
		}
	}

	return false
}

func openRepository(repo configfile.Repository, status *statusList) (*git.Repository, error) {
	switch repository, err := git.PlainOpen(repo.Directory); {

	// If we get ErrRepositoryNotExists here, it means the repo is broken
	case errors.Is(err, git.ErrRepositoryNotExists):
		status.append(repo.Directory, util.CheckColors(color.RedString, "broken"))
		return nil, err

	case err != nil:
		status.appendError(repo.Directory, err)
		return nil, err

	default:
		return repository, nil
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

func updateConfigFlags() {
	var conf *configfile.Configuration
	if configfile.ConfigurationExists() {
		conf = configfile.Load()
	}

	if conf != nil {
		configFlags = conf
	}
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

	var status statusList
	for _, f := range files {
		if !isRepoDir(f, conf.Repositories) {
			status.append(f, color.RedString("untracked"))
		}
	}

	status.print()

	return nil
}

func updateRepoConfig(conf *configfile.Configuration, host string, repository *git.Repository) error {
	repoConf, err := repository.Config()
	if err != nil {
		return err
	}

	section := repoConf.Raw.Section("user")
	profilesMap := conf.Profiles.ToMap()
	section.SetOption("name", profilesMap[host].Fullname)
	section.SetOption("email", profilesMap[host].Email)

	if err := repoConf.Validate(); err != nil {
		return err
	}

	if err := repository.Storer.SetConfig(repoConf); err != nil {
		return err
	}

	return nil
}
