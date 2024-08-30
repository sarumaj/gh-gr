package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	color "github.com/fatih/color"
	configfile "github.com/sarumaj/gh-gr/v2/pkg/configfile"
	restclient "github.com/sarumaj/gh-gr/v2/pkg/restclient"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	logrus "github.com/sirupsen/logrus"
	cobra "github.com/spf13/cobra"
	pool "gopkg.in/go-playground/pool.v3"
)

// prFlags represents the flags for pr command
var prFlags struct {
	base            string
	closedInLast    time.Duration
	closedAfterLast time.Duration
	customQuery     string
	head            string
	state           string
	assignees       []string
	authors         []string
	filters         []string
	labels          []string
	titles          []string
}

// prCmd represents the pr command
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
		Example: "gh gr pr --state open",
		Run: func(*cobra.Command, []string) {
			c := util.Console()
			if !configfile.ConfigurationExists() {
				util.PrintlnAndExit(c.CheckColors(color.RedString, configfile.ConfigNotFound))
			}

			conf := configfile.Load()

			var list configfile.PullRequestList
			listPullRequests(conf, buildPullSearchQuery(), &list, true)

			if len(list) == 0 {
				util.PrintlnAndExit(c.CheckColors(color.RedString, "No pull requests matching following constraints found"))
			}
		},
	}

	flags := prCmd.Flags()
	flags.StringVar(&prFlags.state, "state", "open", "Filter pull requests by state (\"open\", \"closed\", \"all\")")

	flags = prCmd.PersistentFlags()
	flags.StringVar(&prFlags.base, "base", "", "Filter pull requests by base branch")
	flags.DurationVar(&prFlags.closedInLast, "closed-in-last", 0, "Filter pull requests closed in the last time window")
	flags.DurationVar(&prFlags.closedAfterLast, "closed-after-last", 0, "Filter pull requests closed after the last time window")
	flags.StringVar(&prFlags.customQuery, "query", "", "Custom query to filter pull requests")
	flags.StringVar(&prFlags.head, "head", "", "Filter pull requests by head user or head org in the format \"user:ref-name\" or \"organization:ref-name\"")
	flags.StringArrayVar(&prFlags.assignees, "assignee", []string{}, "Glob pattern(s) to filter pull request assignees")
	flags.StringArrayVar(&prFlags.authors, "author", []string{}, "Glob pattern(s) to filter pull request authors")
	flags.StringArrayVar(&prFlags.filters, "match", []string{}, "Glob pattern(s) to filter pull request repositories")
	flags.StringArrayVar(&prFlags.labels, "label", []string{}, "Glob pattern(s) to filter pull request labels")
	flags.StringArrayVar(&prFlags.titles, "title", []string{}, "Regular expression(s) to filter pull request titles")

	prCmd.AddCommand(prCloseCmd, prReopenCmd)

	return prCmd
}()

// pullRequestAction represents a singular action on a pull request.
type pullRequestAction func(*restclient.RESTClient) func(context.Context, string, string, int) error

// buildPullSearchQuery builds a search query for pull requests.
func buildPullSearchQuery() string {
	var fragments []string

	if prFlags.base != "" {
		fragments = append(fragments, fmt.Sprintf("base:%s", prFlags.base))
	}

	if prFlags.head != "" {
		fragments = append(fragments, fmt.Sprintf("head:%s", prFlags.head))
	}

	if prFlags.closedInLast > 0 {
		fragments = append(fragments, fmt.Sprintf("closed:>=%s", time.Now().Add(-prFlags.closedInLast).Format("2006-01-02T15:04:05Z")))
	}

	if prFlags.closedAfterLast > 0 {
		fragments = append(fragments, fmt.Sprintf("closed:<=%s", time.Now().Add(-prFlags.closedAfterLast).Format("2006-01-02T15:04:05Z")))
	}

	for _, assignee := range prFlags.assignees {
		if util.IsGlobMatch(assignee) {
			continue
		}
		fragments = append(fragments, fmt.Sprintf("assignee:%s", assignee))
	}

	for _, author := range prFlags.authors {
		if util.IsGlobMatch(author) {
			continue
		}
		fragments = append(fragments, fmt.Sprintf("author:%s", author))
	}

	for _, label := range prFlags.labels {
		if util.IsGlobMatch(label) {
			continue
		}
		fragments = append(fragments, fmt.Sprintf("label:%s", label))
	}

	for _, title := range prFlags.titles {
		if util.IsGlobMatch(title) || util.IsRegex(title) {
			continue
		}
		fragments = append(fragments, fmt.Sprintf("%s in:title", title))
	}

	if prFlags.state != "" {
		fragments = append(fragments, fmt.Sprintf("state:%s", prFlags.state))
	}

	if prFlags.customQuery != "" {
		fragments = append(fragments, prFlags.customQuery)
	}

	return strings.Join(fragments, " ")
}

// listPullRequests initializes pull requests.
func listPullRequests(conf *configfile.Configuration, filter string, list *configfile.PullRequestList, flush bool) {
	operationLoop[configfile.Repository](prListOperation, "PRs list", operationContextMap{
		"filter": filter,
		"cache":  make(map[string]*restclient.RESTClient),
		"list":   list,
		"keep": func(pull configfile.PullRequest) bool {
			switch {
			case
				len(prFlags.assignees) > 0 && !util.PatternList(prFlags.assignees).GlobMatchAny(pull.Assignees...),
				len(prFlags.filters) > 0 && !util.PatternList(prFlags.filters).GlobMatch(pull.Repository),
				len(prFlags.titles) > 0 && !util.PatternList(prFlags.titles).RegexMatch(pull.Title, conf.Timeout),
				len(prFlags.authors) > 0 && !util.PatternList(prFlags.authors).GlobMatch(pull.Author),
				len(prFlags.labels) > 0 && !util.PatternList(prFlags.labels).GlobMatchAny(pull.Labels...),
				prFlags.closedInLast > 0 && time.Since(pull.ClosedAt) > prFlags.closedInLast,
				prFlags.closedAfterLast > 0 && time.Since(pull.ClosedAt) < prFlags.closedAfterLast:

				return false
			}
			return true
		},
	}, []string{"Title", "Number", "State", "Repository", "Author", "Assignees", "Labels"}, flush)

	if len(*list) == 0 {
		util.PrintlnAndExit(util.Console().CheckColors(color.RedString, "No pull requests found"))
	}
}

func prDoOperation(_ pool.WorkUnit, args operationContext) {
	conf := unwrapOperationContext[*configfile.Configuration](args, "conf")
	pr := unwrapOperationContext[configfile.PullRequest](args, "object")
	status := unwrapOperationContext[*operationStatus](args, "status")
	cache := unwrapOperationContext[map[string]*restclient.RESTClient](args, "cache")
	action := unwrapOperationContext[pullRequestAction](args, "action")
	newState := unwrapOperationContext[string](args, "newState")

	logger := loggerEntry.WithField("command", "pr").WithField("target_state", newState).WithField("repository", pr.Repository)

	if pr.State == newState {
		status.appendRow(pr.Title, pr.Number, pr.Status(), pr.Author, pr.Assignees, pr.Labels)
		return
	}

	host := util.GetHostnameFromPath(pr.URL)
	client, ok := cache[host]

	if !ok {
		token, ok := configfile.GetTokens()[host]
		if !ok {
			logger.Warnf("Failed to retrieve token for host: %q", host)
			pr.Error = configfile.PullRequestError("failed to retrieve token")
			status.appendRow(pr.Title, pr.Number, pr.Status(), pr.Author, pr.Assignees, pr.Labels)
			return
		}

		var err error
		client, err = restclient.NewRESTClient(conf, restclient.ClientOptions{
			AuthToken:   token,
			Log:         logger.WriterLevel(logrus.DebugLevel),
			LogColorize: util.Console().ColorsEnabled(),
			Host:        host,
		}, globalNonPersistentFlags.retry)
		if err != nil {
			logger.Warnf("Failed to create REST client: %v", err)
			pr.Error = configfile.PullRequestError("failed to retrieve token")
			status.appendRow(pr.Title, pr.Number, pr.Status(), pr.Author, pr.Assignees, pr.Labels)
			return
		}

		cache[host] = client
	}

	owner, repo, _ := strings.Cut(pr.Repository, "/")
	if err := action(client)(args.Context, owner, repo, pr.Number); err != nil {
		prefix, content, ok := strings.Cut(err.Error(), ": ")
		if ok {
			pr.Error = configfile.PullRequestError(content)
		} else {
			pr.Error = configfile.PullRequestError(prefix)
		}

	} else {
		pr.State = newState
	}

	status.appendRow(pr.Title, pr.Number, pr.Status(), pr.Author, pr.Assignees, pr.Labels)
}

func prListOperation(_ pool.WorkUnit, args operationContext) {
	conf := unwrapOperationContext[*configfile.Configuration](args, "conf")
	repo := unwrapOperationContext[configfile.Repository](args, "object")
	status := unwrapOperationContext[*operationStatus](args, "status")
	keep := unwrapOperationContext[func(configfile.PullRequest) bool](args, "keep")
	filter := unwrapOperationContext[string](args, "filter")
	cache := unwrapOperationContext[map[string]*restclient.RESTClient](args, "cache")
	list := unwrapOperationContext[*configfile.PullRequestList](args, "list")

	logger := loggerEntry.WithField("command", "pr").WithField("repository", repo.Directory)

	host := util.GetHostnameFromPath(repo.URL)
	client, ok := cache[host]
	if !ok {
		token, ok := configfile.GetTokens()[host]
		if !ok {
			logger.Warnf("Failed to retrieve token for host: %q", host)
			status.appendRow("", "", fmt.Errorf("failed to retrieve token"), repo.Directory, "", []string{}, []string{})
			return
		}

		var err error
		client, err = restclient.NewRESTClient(conf, restclient.ClientOptions{
			AuthToken:   token,
			Log:         logger.WriterLevel(logrus.DebugLevel),
			LogColorize: util.Console().ColorsEnabled(),
			Host:        host,
		}, globalNonPersistentFlags.retry)
		if err != nil {
			logger.Warnf("Failed to create REST client: %v", err)
			status.appendRow("", "", err, repo.Directory, "", []string{}, []string{})
			return
		}

		cache[host] = client
	}

	slug := configfile.GetRepositorySlugFromURL(repo)
	owner, repoName, _ := strings.Cut(slug, "/")

	pulls, err := client.GetOrgRepoPulls(args.Context, owner, repoName, filter)
	if err != nil {
		status.appendRow("", "", err, repo.Directory, "", []string{}, []string{})
		return
	}

	for _, pr := range pulls {
		entry := configfile.PullRequestFromResponse(pr)
		if !keep(entry) {
			continue
		}

		status.appendRow(entry.Title, entry.Number, entry.Status(), entry.Repository, entry.Author, entry.Assignees, entry.Labels)
		list.Append(entry)
	}
}
