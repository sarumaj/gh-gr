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
			if configFlags == nil {
				configFlags = &configfile.Configuration{}
			}
			runInit(configFlags, false)
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
	switch exists := configfile.ConfigurationExists(); {

	case exists && !update:
		fmt.Fprintln(os.Stderr, util.CheckColors(color.RedString, configfile.ConfigShouldNotExist))
		return

	case !exists && update:
		fmt.Fprintln(os.Stderr, util.CheckColors(color.RedString, configfile.ConfigNotFound))
		return

	case conf == nil:
		conf = configfile.Load()

	}

	logger := util.Logger()
	if conf.Verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	client, err := restclient.NewRESTClient(conf, restclient.ClientOptions{
		AuthToken:   conf.GetToken(),
		Log:         logger.WriterLevel(logrus.DebugLevel),
		LogColorize: true,
		Host:        conf.BaseURL,
	})
	util.FatalIfError(err)

	ctx := context.Background()
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

	repos, err := client.GetUserRepos(ctx)
	util.FatalIfError(err)

	orgs, err := client.GetUserOrgs(ctx)
	util.FatalIfError(err)

	for _, org := range orgs {
		orgRepos, err := client.GetOrgRepos(ctx, org.Login)
		util.FatalIfError(err)
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

			continue
		}

		dir := repo.FullName
		if !conf.SubDirectories {
			dir = strings.ReplaceAll(dir, "/", "_")
			dir = strings.ReplaceAll(dir, conf.Username+"_", "")
		}

		conf.Repositories = append(conf.Repositories, configfile.Repository{
			URL:       repo.CloneURL,
			Branch:    repo.DefaultBranch,
			ParentURL: repo.Parent.CloneURL,
			Directory: filepath.Join(conf.BaseDirectory, dir),
		})
	}

	if err := addGitAliases(); err != nil {
		logger.Warnf("failed to set up git alias commands: %v", err)
	}

	conf.Save()
}
