package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	restclient "github.com/sarumaj/gh-gr/pkg/restclient"
	util "github.com/sarumaj/gh-gr/pkg/util"
	logrus "github.com/sirupsen/logrus"
	cobra "github.com/spf13/cobra"
)

var initCmd = func() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize repository mirror",
		Run: func(*cobra.Command, []string) {
			// call copy to initialize all empty config fields
			runInit(configFlags.Copy(), false)
		},
		PostRun: func(*cobra.Command, []string) {
			updateConfigFlags()
		},
	}

	flags := initCmd.Flags()
	flags.StringVarP(&configFlags.BaseDirectory, "dir", "d", ".", "Directory in which repositories will be stored (either absolute or relative)")
	flags.BoolVarP(&configFlags.SubDirectories, "subdirs", "s", false, "Enable creation of separate subdirectories for each org/user")
	flags.StringArrayVarP(&configFlags.Excluded, "exclude", "e", []string{}, "Regular expressions of repositories to exclude")
	flags.StringArrayVarP(&configFlags.Included, "include", "i", []string{}, "Regular expressions of repositories to include (bind stronger than exclusion list)")

	return initCmd
}()

func runInit(conf *configfile.Configuration, update bool) {
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

	util.PathSanitize(&conf.BaseDirectory)
	if filepath.IsAbs(conf.BaseDirectory) {
		conf.AbsoluteDirectoryPath = filepath.Dir(conf.BaseDirectory)
		conf.BaseDirectory = filepath.Base(conf.BaseDirectory)

	} else {
		abs, err := filepath.Abs(conf.BaseDirectory)
		util.FatalIfError(err)
		conf.AbsoluteDirectoryPath = filepath.Dir(abs)

	}

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

		rate, err := client.GetRateLimit(ctx)
		util.FatalIfError(err)
		restclient.CheckRateLimit(rate)

		user, err := client.GetUser(ctx)
		util.FatalIfError(err)

		profile := &configfile.Profile{
			Username: user.Login,
			Fullname: user.Name,
			Email:    user.Email,
			Host:     host,
		}

		if profile.Email == "" {
			profile.Email = fmt.Sprintf("%d-%s@users.noreply.github.com", user.ID, user.Login)
		}

		conf.Profiles = append(conf.Profiles, *profile)

		logger.Debugf("Username: %s, name: %s, email: %s", profile.Username, profile.Fullname, profile.Email)

		repos, err := client.GetUserRepos(ctx)
		util.FatalIfError(err)
		logger.Debugf("Retrieved %d user repositories", len(repos))

		orgs, err := client.GetUserOrgs(ctx)
		util.FatalIfError(err)
		logger.Debugf("Retrieved %d user organizations", len(orgs))

		for _, org := range orgs {
			orgRepos, err := client.GetOrgRepos(ctx, org.Login)
			util.FatalIfError(err)
			logger.Debugf("Retrieved %d repositories for organization: %s", len(orgRepos), org.Login)
			repos = append(repos, orgRepos...)
		}

		for _, repo := range repos {
			switch {
			case
				// not explicitly included
				len(conf.Included) > 0 && !util.RegexList(conf.Included).Match(repo.FullName),

				// explicitly excluded and not included
				util.RegexList(conf.Excluded).Match(repo.FullName) && !util.RegexList(conf.Included).Match(repo.FullName):

				logger.Debugf("Excluding repository: %s", repo.FullName)
				continue
			}

			dir := repo.FullName
			if !conf.SubDirectories {
				dir = strings.ReplaceAll(dir, "/", "_")
				dir = strings.ReplaceAll(dir, profile.Username+"_", "")
			}
			dir = filepath.Join(conf.BaseDirectory, filepath.FromSlash(dir))
			util.PathSanitize(&dir)

			logger.Debugf("Adding repository: %s", dir)
			conf.Repositories.Append(configfile.Repository{
				URL:       repo.CloneURL,
				Branch:    repo.DefaultBranch,
				ParentURL: repo.Parent.CloneURL,
				Directory: dir,
			})
		}

		if err := addGitAliases(); err != nil {
			logger.Warnf("failed to set up git alias commands: %v", err)
		}
	}

	conf.Save()
}
