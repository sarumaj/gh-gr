package commands

import (
	"context"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/v2/pkg/configfile"
	"github.com/sarumaj/gh-gr/v2/pkg/restclient"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	cobra "github.com/spf13/cobra"
)

// prReopenCmd represents the pr command
var prReopenCmd = func() *cobra.Command {
	prReopenCmd := &cobra.Command{
		Use:     "reopen",
		Short:   "Reopen closed pull requests",
		Long:    "Reopen closed pull requests.",
		Example: "gh gr pr reopen",
		Run: func(*cobra.Command, []string) {
			c := util.Console()
			if !configfile.ConfigurationExists() {
				util.PrintlnAndExit("%s", c.CheckColors(color.RedString, configfile.ConfigNotFound))
			}

			conf := configfile.Load()

			var list configfile.PullRequestList
			listPullRequests(conf, buildPullSearchQuery(), &list, true)

			if len(list) == 0 {
				util.PrintlnAndExit("%s", c.CheckColors(color.RedString, "No pull requests matching provided constrains found"))
			}

			operationLoop(prDoOperation, "Reopen", operationContextMap{
				"cache": make(map[string]*restclient.RESTClient),
				"action": pullRequestAction(func(client *restclient.RESTClient) func(context.Context, string, string, int) error {
					return client.ReopenPullRequest
				}),
				"newState": "open",
				"headers":  []string{"Title", "Number", "State", "Repository", "Author", "Assignees", "Labels"},
			}, list...)
		},
	}

	return prReopenCmd
}()
