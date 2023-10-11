package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	color "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	extras "github.com/sarumaj/gh-gr/pkg/extras"
	restclient "github.com/sarumaj/gh-gr/pkg/restclient"
	util "github.com/sarumaj/gh-gr/pkg/util"
	logrus "github.com/sirupsen/logrus"
)

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

func changeProgressbarText(bar *util.Progressbar, conf *configfile.Configuration, verb string, repo configfile.Repository) {
	if bar != nil && conf != nil {
		bar.Describe(util.CheckColors(color.BlueString, conf.GetProgressbarDescriptionForVerb(verb, repo)))
	}
}

func initializeOrUpdateConfig(conf *configfile.Configuration, update bool) {
	var logger *logrus.Entry
	if update {
		logger = loggerEntry.WithField("command", "update")
	} else {
		logger = loggerEntry.WithField("command", "init")
	}

	exists := configfile.ConfigurationExists()
	logger.Debugf("Exists: %t, update: %t, conf: %t", exists, update, conf != nil)

	switch {

	case exists && !update:
		util.PrintlnAndExit(util.CheckColors(color.RedString, configfile.ConfigShouldNotExist))

	case !exists && update:
		util.PrintlnAndExit(util.CheckColors(color.RedString, configfile.ConfigNotFound))

	case exists && conf == nil:
		conf = configfile.Load()

	case conf == nil:
		util.PrintlnAndExit(util.CheckColors(color.RedString, configfile.ConfigNotFound))

	}

	conf.SanitizeDirectory()

	tokens := configfile.GetTokens()
	logger.Debugf("Retrieved tokens: %d", len(tokens))

	defer util.PreventInterrupt()()

	for host, token := range tokens {
		client, err := restclient.NewRESTClient(conf, restclient.ClientOptions{
			AuthToken:   token,
			Log:         logger.WriterLevel(logrus.DebugLevel),
			LogColorize: true,
			Host:        host,
		})
		util.FatalIfError(err)

		ctx, cancel := context.WithTimeout(context.Background(), conf.Timeout)
		defer cancel()

		user, err := client.GetUser(ctx)
		util.FatalIfError(err)

		profile := configfile.NewProfile(user, host)
		conf.Profiles.Append(profile)
		logger.Debugf("Username: %s, name: %s, email: %s", profile.Username, profile.Fullname, profile.Email)

		repos, err := client.GetAllUserRepos(ctx)
		util.FatalIfError(err)
		logger.Debugf("Retrieved %d user repositories", len(repos))

		conf.FilterRepositories(&repos)
		logger.Debugf("Applied filters: %d repositories remaining", len(repos))

		conf.AppendRepositories(user, repos...)

		if err := addGitAliases(); err != nil {
			logger.Debugf("failed to set up git alias commands: %v", err)
		}
	}

	conf.Save()
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

func openRepository(repo configfile.Repository, status *operationStatus) (*git.Repository, error) {
	switch repository, err := git.PlainOpen(repo.Directory); {

	// If we get ErrRepositoryNotExists here, it means the repo is broken
	case errors.Is(err, git.ErrRepositoryNotExists):
		status.appendErrorRow(repo.Directory, fmt.Errorf("broken"))
		return nil, err

	case err != nil:
		status.appendErrorRow(repo.Directory, err)
		return nil, err

	default:
		return repository, nil
	}
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
