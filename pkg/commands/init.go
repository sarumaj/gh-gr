package commands

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/sarumaj/gh-pr/pkg/configfile"
	"github.com/sarumaj/gh-pr/pkg/restclient"
	"github.com/sarumaj/gh-pr/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var _ = func() *cobra.Command {
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

	initCmd.Flags().StringVarP(&configFlags.BaseURL, "url", "r", host, "GitHub (Enterprise) URL")
	initCmd.Flags().StringVarP(&configFlags.BaseDirectory, "dir", "d", ".", "Directory in which repositories will be stored")
	initCmd.Flags().BoolVarP(&configFlags.SubDirectories, "subdirs", "s", false, "Enable creation of separate subdirectories for each org/user")
	initCmd.Flags().StringSliceVarP(&configFlags.Excluded, "exclude", "e", []string{}, "Regular expressions of repositories to exclude")
	initCmd.Flags().StringSliceVarP(&configFlags.Included, "include", "i", []string{}, "Regular expressions of repositories to include (bind stronger than exclusion list)")

	rootCmd.AddCommand(initCmd)

	return initCmd
}()

func runInit(conf *configfile.Configuration, update bool) {
	if conf.Exists() && !update {
		util.FatalIfError(configfile.ErrConfExists)
	}

	logger := util.Logger()
	if conf.Verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	token, _ := auth.TokenForHost(conf.BaseURL)
	client, err := restclient.NewRESTClient(conf, restclient.ClientOptions{
		AuthToken:   token,
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
	conf.Email = user.Email

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
		if true &&
			util.RegexList(conf.Excluded).Match(repo.FullName) &&
			!util.RegexList(conf.Included).Match(repo.FullName) {

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

	conf.Save()
}
