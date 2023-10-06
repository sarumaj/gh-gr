package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	auth "github.com/cli/go-gh/v2/pkg/auth"
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
		Run: func(cmd *cobra.Command, args []string) {
			runInit(configFlags.Copy(), false)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			updateConfigFlags()
		},
	}

	host, _ := auth.DefaultHost()

	flags := initCmd.Flags()
	flags.StringVarP(&configFlags.BaseURL, "url", "u", host, "GitHub (Enterprise) URL")
	flags.StringVarP(&configFlags.BaseDirectory, "dir", "d", ".", "Directory in which repositories will be stored")
	flags.BoolVarP(&configFlags.SubDirectories, "subdirs", "s", false, "Enable creation of separate subdirectories for each org/user")
	flags.StringArrayVarP(&configFlags.Excluded, "exclude", "e", []string{}, "Regular expressions of repositories to exclude")
	flags.StringArrayVarP(&configFlags.Included, "include", "i", []string{}, "Regular expressions of repositories to include (bind stronger than exclusion list)")

	return initCmd
}()

func runInit(conf *configfile.Configuration, update bool) {
	interrupt := util.NewInterrupt()
	defer interrupt.Stop()

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
		fmt.Fprintln(os.Stderr, util.CheckColors(color.RedString, configfile.ConfigShouldNotExist))
		return

	case !exists && update:
		fmt.Fprintln(os.Stderr, util.CheckColors(color.RedString, configfile.ConfigNotFound))
		return

	}

	if exists && conf == nil {
		conf = configfile.Load()

	} else if conf == nil {
		fmt.Fprintln(os.Stderr, util.CheckColors(color.RedString, configfile.ConfigNotFound))
		return

	}

	client, err := restclient.NewRESTClient(conf, restclient.ClientOptions{
		AuthToken:   conf.GetToken(),
		Log:         logger.WriterLevel(logrus.DebugLevel),
		LogColorize: true,
		Host:        conf.BaseURL,
	})
	util.FatalIfError(err)

	ctx, cancel := context.WithTimeout(context.Background(), conf.Timeout)
	defer cancel()

	rate, err := client.GetRateLimit(ctx)
	util.FatalIfError(err)
	util.FatalIfError(restclient.CheckRateLimit(rate))

	user, err := client.GetUser(ctx)
	util.FatalIfError(err)

	conf.Username = user.Login
	conf.Fullname = user.Name

	if user.Email == "" {
		conf.Email = fmt.Sprintf("%d-%s@users.noreply.github.com", user.ID, user.Login)
	} else {
		conf.Email = user.Email
	}

	logger.Debugf("Username: %s, name: %s, email: %s", conf.Username, conf.Fullname, conf.Email)

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
			dir = strings.ReplaceAll(dir, conf.Username+"_", "")
		}
		dir = filepath.ToSlash(filepath.Join(conf.BaseDirectory, filepath.FromSlash(dir)))

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

	conf.Save()
}
