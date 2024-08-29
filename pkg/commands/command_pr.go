package commands

import (
	"context"
	"strings"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/v2/pkg/configfile"
	restclient "github.com/sarumaj/gh-gr/v2/pkg/restclient"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	supererrors "github.com/sarumaj/go-super/errors"
	logrus "github.com/sirupsen/logrus"
	cobra "github.com/spf13/cobra"
)

// prFlags represents the flags for pr command
var prFlags struct {
	base      string
	head      string
	state     string
	sort      string
	assignees []string
	authors   []string
	filters   []string
	labels    []string
	titles    []string
	close     bool
}

// initCmd represents the init command
var prCmd = func() *cobra.Command {
	prCmd := &cobra.Command{
		Use:     "pr",
		Aliases: []string{"pulls", "prs"},
		Short:   "List and modify pull requests",
		Long: "List and modify pull requests.\n\n" +
			"Supports listing pull requests for a given user and filtering by glob match and regular expressions.\n" +
			"Regular expressions support following features:\n\n" +
			"\t- Python-style capture groups (?P<name>re)\n" +
			"\t- .NET-style capture groups (?<name>re) or (?'name're)\n" +
			"\t- comments (?#comment)\n" +
			"\t- possessive match (?>re)\n" +
			"\t- positive lookahead (?=re)\n" +
			"\t- negative lookahead (?!re)\n" +
			"\t- positive lookbehind (?<=re)\n" +
			"\t- negative lookbehind (?<!re)\n" +
			"\t- back reference \\1\n" +
			"\t- named back reference \\k'name'\n" +
			"\t- named ascii character class [[:foo:]]\n" +
			"\t- conditionals (?(expr)yes|no)\n",
		Example: "gh gr pr " +
			"--state open",
		Run: func(*cobra.Command, []string) {
			c := util.Console()
			if !configfile.ConfigurationExists() {
				util.PrintlnAndExit(c.CheckColors(color.RedString, configfile.ConfigNotFound))
			}

			logger := loggerEntry.WithField("command", "pr")
			conf := configfile.Load()

			pulls, err := listPullRequests(conf, logger, map[string]string{
				"state": prFlags.state,
				"sort":  prFlags.sort,
				"base":  prFlags.base,
				"head":  prFlags.head,
			})
			supererrors.Except(err)

			if len(pulls) == 0 {
				util.PrintlnAndExit(c.CheckColors(color.RedString, "No pull requests found"))
			}

			var filtered []configfile.PullRequest
			for _, pull := range pulls {
				switch {
				case
					len(prFlags.assignees) > 0 && !util.PatternList(prFlags.assignees).GlobMatchAny(pull.Assignees...),
					len(prFlags.filters) > 0 && !util.PatternList(prFlags.filters).GlobMatch(pull.Repository),
					len(prFlags.titles) > 0 && !util.PatternList(prFlags.titles).RegexMatch(pull.Title, conf.Timeout),
					len(prFlags.authors) > 0 && !util.PatternList(prFlags.authors).RegexMatch(pull.Author, conf.Timeout),
					len(prFlags.labels) > 0 && !util.PatternList(prFlags.labels).RegexMatchAny(conf.Timeout, pull.Labels...):

					continue
				}

				filtered = append(filtered, pull)
			}

			if len(filtered) == 0 {
				util.PrintlnAndExit(c.CheckColors(color.RedString, "No pull requests matching provided constrains found"))
			}

			if prFlags.close {
				filtered = closePullRequests(conf, logger, filtered)
			}

			status := newOperationStatus()
			for _, pr := range filtered {
				if len(pr.Title) > 30 {
					pr.Title = pr.Title[:27] + "..."
				}
				status.appendRow(pr.Title, pr.Number, pr.State, pr.Repository, pr.Author,
					strings.Join(pr.Assignees, ","),
					strings.Join(pr.Labels, ","))
			}

			status.Sort().Print()
		},
	}

	flags := prCmd.Flags()
	flags.StringVar(&prFlags.state, "state", "open", "Filter pull requests by state (\"open\", \"closed\", \"all\")")
	flags.StringVar(&prFlags.sort, "sort", "created", "Sort pull requests by field (\"created\", \"updated\", \"popularity\", \"long-running\")")
	flags.StringVar(&prFlags.base, "base", "", "Filter pull requests by base branch")
	flags.StringVar(&prFlags.head, "head", "", "Filter pull requests by head user or head org in the format \"user:ref-name\" or \"organization:ref-name\"")
	flags.BoolVar(&prFlags.close, "close", false, "Close pull requests")
	flags.StringArrayVar(&prFlags.assignees, "assignee", []string{}, "Glob pattern(s) to filter pull request assignees")
	flags.StringArrayVar(&prFlags.authors, "author", []string{}, "Regular expression(s) to filter pull request authors")
	flags.StringArrayVar(&prFlags.filters, "match", []string{}, "Glob pattern(s) to filter pull request repositories")
	flags.StringArrayVar(&prFlags.labels, "label", []string{}, "Regular expression(s) to filter pull request labels")
	flags.StringArrayVar(&prFlags.titles, "title", []string{}, "Regular expression(s) to filter pull request titles")

	return prCmd
}()

// closePullRequests closes pull requests.
func closePullRequests(conf *configfile.Configuration, logger *logrus.Entry, prs []configfile.PullRequest) (out []configfile.PullRequest) {
	tokens := configfile.GetTokens()
	logger.Debugf("Retrieved tokens: %d", len(tokens))

	ctx, cancel := context.WithTimeout(context.Background(), conf.Timeout)
	defer cancel()

	cache := make(map[string]*restclient.RESTClient)
	for _, pr := range prs {
		host := util.GetHostnameFromPath(pr.URL)
		client, ok := cache[host]
		if !ok {
			token, ok := tokens[host]
			if !ok {
				logger.Warnf("Failed to retrieve token for host: %q", host)
				pr.State = "failed"
				out = append(out, pr)
				continue
			}

			client, err := restclient.NewRESTClient(conf, restclient.ClientOptions{
				AuthToken:   token,
				Log:         logger.WriterLevel(logrus.DebugLevel),
				LogColorize: util.Console().ColorsEnabled(),
				Host:        host,
			})
			if err != nil {
				util.PrintlnAndExit("Failed to create REST client: %v", err)
			}

			cache[host] = client
		}

		owner, repo, ok := strings.Cut(pr.Repository, "/")
		if !ok {
			logger.Warnf("Failed to parse repository: %q", pr.Repository)
			pr.State = "failed"
			out = append(out, pr)
			continue
		}

		if err := client.ClosePullRequest(ctx, owner, repo, pr.Number); err != nil {
			logger.Warnf("Failed to close pull request: %v", err)
			pr.State = "failed"
			out = append(out, pr)
		}
	}

	return
}

// listPullRequest lists pull requests.
func listPullRequests(conf *configfile.Configuration, logger *logrus.Entry, filters map[string]string) ([]configfile.PullRequest, error) {
	tokens := configfile.GetTokens()
	logger.Debugf("Retrieved tokens: %d", len(tokens))

	ctx, cancel := context.WithTimeout(context.Background(), conf.Timeout)
	defer cancel()

	var pulls []configfile.PullRequest
	for host, token := range tokens {
		client, err := restclient.NewRESTClient(conf, restclient.ClientOptions{
			AuthToken:   token,
			Log:         logger.WriterLevel(logrus.DebugLevel),
			LogColorize: util.Console().ColorsEnabled(),
			Host:        host,
		})
		if err != nil {
			return nil, err
		}

		got, err := client.GetAllUserPulls(ctx, conf.Included, conf.Excluded, filters)
		if err != nil {
			return nil, err
		}

		for _, pr := range got {
			entry := configfile.PullRequest{
				Number:     pr.Number,
				Title:      pr.Title,
				URL:        pr.URL,
				State:      pr.State,
				Author:     pr.User.Login,
				Assignees:  []string{pr.Assignee.Login},
				Repository: pr.Owner + "/" + pr.Repository,
			}

			for _, assignee := range pr.Assignees {
				entry.Assignees = append(entry.Assignees, assignee.Login)
			}

			for _, label := range pr.Labels {
				entry.Labels = append(entry.Labels, label.Name)
			}

			pulls = append(pulls, entry)
		}
	}

	return pulls, nil
}
