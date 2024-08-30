package commands

import (
	"context"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/v2/pkg/configfile"
	restclient "github.com/sarumaj/gh-gr/v2/pkg/restclient"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	cobra "github.com/spf13/cobra"
)

// prCloseCmd represents the pr command
var prCloseCmd = func() *cobra.Command {
	prCloseCmd := &cobra.Command{
		Use:     "close",
		Short:   "Close open pull requests",
		Long:    "Close open pull requests.",
		Example: "gh gr pr close",
		Run: func(*cobra.Command, []string) {
			c := util.Console()
			if !configfile.ConfigurationExists() {
				util.PrintlnAndExit(c.CheckColors(color.RedString, configfile.ConfigNotFound))
			}

			conf := configfile.Load()

			var list configfile.PullRequestList
			listPullRequests(conf, buildPullSearchQuery(), &list, false)

			if len(list) == 0 {
				util.PrintlnAndExit(c.CheckColors(color.RedString, "No pull requests matching provided constrains found"))
			}

			operationLoop(prDoOperation, "Close", operationContextMap{
				"cache": make(map[string]*restclient.RESTClient),
				"action": pullRequestAction(func(client *restclient.RESTClient) func(context.Context, string, string, int) error {
					return client.ClosePullRequest
				}),
				"newState": "closed",
			}, []string{"Title", "Number", "State", "Repository", "Author", "Assignees", "Labels"}, true, list...)
		},
	}

	return prCloseCmd
}()
