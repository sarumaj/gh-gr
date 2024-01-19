package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	color "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	extras "github.com/sarumaj/gh-gr/pkg/extras"
	restclient "github.com/sarumaj/gh-gr/pkg/restclient"
	util "github.com/sarumaj/gh-gr/pkg/util"
	supererrors "github.com/sarumaj/go-super/errors"
	logrus "github.com/sirupsen/logrus"
)

// Write git alias commands into local repository config file.
func addGitAliases() error {
	var ga []struct {
		Alias   string `json:"alias"`
		Command string `json:"command"`
	}
	if err := json.Unmarshal(extras.GitAliasesJSON, &ga); err != nil {
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

// Change progressbar description for given repository.
func changeProgressbarText(bar *util.Progressbar, conf *configfile.Configuration, verb string, repo configfile.Repository) {
	if bar != nil && conf != nil {
		c := util.Console()
		bar.Describe(c.CheckColors(color.BlueString, conf.GetProgressbarDescriptionForVerb(verb, repo)))
	}
}

// Initialize new configuration or update existing one.
func initializeOrUpdateConfig(conf *configfile.Configuration, update bool) {
	var logger *logrus.Entry
	if update {
		logger = loggerEntry.WithField("command", "update")
	} else {
		logger = loggerEntry.WithField("command", "init")
	}

	exists := configfile.ConfigurationExists()
	logger.Debugf("Exists: %t, update: %t, conf: %t", exists, update, conf != nil)

	c := util.Console()
	switch {

	case exists && !update:
		util.PrintlnAndExit(c.CheckColors(color.RedString, configfile.ConfigShouldNotExist))

	case !exists && update:
		util.PrintlnAndExit(c.CheckColors(color.RedString, configfile.ConfigNotFound))

	}

	switch {

	case exists && conf == nil:
		conf = configfile.Load()

	case conf == nil:
		util.PrintlnAndExit(c.CheckColors(color.RedString, configfile.ConfigNotFound))

	}

	if update {
		conf.Profiles = nil
		conf.Repositories = nil

	} else {
		conf.SanitizeDirectory()

	}

	tokens := configfile.GetTokens()
	logger.Debugf("Retrieved tokens: %d", len(tokens))

	defer util.PreventInterrupt().Stop()
	for host, token := range tokens {
		client, err := restclient.NewRESTClient(conf, restclient.ClientOptions{
			AuthToken:   token,
			Log:         logger.WriterLevel(logrus.DebugLevel),
			LogColorize: true,
			Host:        host,
		})
		supererrors.Except(err)

		ctx, cancel := context.WithTimeout(context.Background(), conf.Timeout)
		defer cancel()

		user, err := client.GetUser(ctx)
		supererrors.Except(err)

		profile := configfile.NewProfile(user, host)
		conf.Profiles.Append(profile)
		logger.Debugf("Username: %s, name: %s, email: %s", profile.Username, profile.Fullname, profile.Email)

		repos, err := client.GetAllUserRepos(ctx, conf.Included, conf.Excluded)
		supererrors.Except(err)
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

// Open local repository.
func openRepository(repo configfile.Repository, status *operationStatus) (*git.Repository, error) {
	switch repository, err := git.PlainOpen(repo.Directory); {

	// If we get ErrRepositoryNotExists here, it means the repo is broken
	case errors.Is(err, git.ErrRepositoryNotExists):
		status.appendRow(repo.Directory, fmt.Errorf("broken"))
		return nil, err

	case err != nil:
		status.appendRow(repo.Directory, err)
		return nil, err

	default:
		return repository, nil
	}
}

// Update configFlags from loaded configuration.
func updateConfigFlags() {
	var conf *configfile.Configuration
	if configfile.ConfigurationExists() {
		conf = configfile.Load()
	}

	if conf != nil {
		configFlags = conf
	}
}
