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
	logger := util.Logger()
	if configFlags.Verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	var entry *logrus.Entry
	if update {
		entry = logger.WithField("command", "update")
	} else {
		entry = logger.WithField("command", "init")
	}

	exists := configfile.ConfigurationExists()
	entry.Debugf("Check configuration: %t, update: %t, conf: %t", exists, update, conf != nil)

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

	entry.Debug("Retrieving rate limitation")
	rate, err := client.GetRateLimit(ctx)
	util.FatalIfError(err)
	util.FatalIfError(restclient.CheckRateLimit(rate))

	entry.Debug("Retrieving user")
	user, err := client.GetUser(ctx)
	util.FatalIfError(err)

	conf.Username = user.Login
	conf.Fullname = user.Name

	if user.Email == "" {
		conf.Email = fmt.Sprintf("%d-%s@users.noreply.github.com", user.ID, user.Login)
	} else {
		conf.Email = user.Email
	}

	entry.Debugf("Retrieved username: %s, name: %s, email: %s", conf.Username, conf.Fullname, conf.Email)

	entry.Debug("Retrieving user repositories")
	repos, err := client.GetUserRepos(ctx)
	util.FatalIfError(err)
	entry.Debugf("Retrieved %d user repositories", len(repos))

	entry.Debug("Retrieving user organizations")
	orgs, err := client.GetUserOrgs(ctx)
	util.FatalIfError(err)
	entry.Debugf("Retrieved %d user organizations", len(orgs))

	for _, org := range orgs {
		entry.Debugf("Retrieving repositories for organization: %s", org.Login)
		orgRepos, err := client.GetOrgRepos(ctx, org.Login)
		util.FatalIfError(err)
		entry.Debugf("Retrieved %d repositories for organization: %s", len(orgRepos), org.Login)
		repos = append(repos, orgRepos...)
	}

	for _, repo := range repos {
		switch {
		case
			// not explicitly included
			len(conf.Included) > 0 &&
				!util.RegexList(conf.Included).Match(repo.FullName),

			// explicitly excluded and not included
			util.RegexList(conf.Excluded).Match(repo.FullName) &&
				!util.RegexList(conf.Included).Match(repo.FullName):

			entry.Debugf("Excluding repository: %s", repo.FullName)
			continue
		}

		dir := repo.FullName
		if !conf.SubDirectories {
			dir = strings.ReplaceAll(dir, "/", "_")
			dir = strings.ReplaceAll(dir, conf.Username+"_", "")
			dir = filepath.Join(conf.BaseDirectory, dir)
		}

		entry.Debugf("Adding repository: %s", dir)
		conf.Repositories.Append(configfile.Repository{
			URL:       repo.CloneURL,
			Branch:    repo.DefaultBranch,
			ParentURL: repo.Parent.CloneURL,
			Directory: dir,
		})
	}

	entry.Debug("Installing git alias commands")
	if err := addGitAliases(); err != nil {
		logger.Warnf("failed to set up git alias commands: %v", err)
	}

	entry.Debug("Saving")
	conf.Save()
}
