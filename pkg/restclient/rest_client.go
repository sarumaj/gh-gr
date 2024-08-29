package restclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	api "github.com/cli/go-gh/v2/pkg/api"
	configfile "github.com/sarumaj/gh-gr/v2/pkg/configfile"
	resources "github.com/sarumaj/gh-gr/v2/pkg/restclient/resources"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
)

// Module logger.
var loggerEntry = util.Logger.WithField("mod", "restclient")

// Ship client options.
type ClientOptions = api.ClientOptions

// REST API client.
type RESTClient struct {
	*api.RESTClient
	*configfile.Configuration
	*util.Progressbar
}

// Close a pull request.
func (c RESTClient) ClosePullRequest(ctx context.Context, owner, repo string, number int) error {
	c.Describe(fmt.Sprintf("Closing pull request %d for GitHub repository: %s/%s...", number, owner, repo))
	return c.DoWithContext(ctx, http.MethodPatch,
		newRequestPath(pullEp.Format(map[string]any{"owner": owner, "repo": repo, "number": number})).String(),
		strings.NewReader(`{"state":"closed"}`), nil)
}

// Implements DoWithContext method.
func (c RESTClient) DoWithContext(ctx context.Context, method string, path string, body io.Reader, response any) error {
	return c.RESTClient.DoWithContext(ctx, method, path, body, response)
}

// Get all pull requests for given user.
func (c RESTClient) GetAllUserPulls(ctx context.Context, include, exclude []string, filter map[string]string) ([]resources.PullRequest, error) {
	repos, err := c.GetAllUserRepos(ctx, include, exclude)
	if err != nil {
		return nil, err
	}

	var pulls []resources.PullRequest
	for _, repo := range repos {
		c.Describe(fmt.Sprintf("Retrieving pull requests for GitHub repository: %s...", repo.FullName))
		pullsRepo, err := c.GetOrgRepoPulls(ctx, repo.Owner.Login, repo.Name, filter)
		if err != nil {
			return nil, err
		}

		pulls = append(pulls, pullsRepo...)
	}

	return pulls, nil
}

// Get all repositories for given user and the organizations he belongs to.
func (c RESTClient) GetAllUserRepos(ctx context.Context, include, exclude []string) ([]resources.Repository, error) {
	repos, err := c.GetUserRepos(ctx)
	if err != nil {
		return nil, err
	}

	orgs, err := c.GetUserOrgs(ctx)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(1<<63 - 1)
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	for _, org := range orgs {
		c.ProgressBar.Describe(fmt.Sprintf("Retrieving repositories for GitHub organization: %s...", org.Login))

		switch includes, excludes := util.PatternList(include), util.PatternList(exclude); {
		case
			len(include) > 0 && !(includes.RegexMatch(org.Login+"/someRepository", timeout) || includes.RegexMatch(org.Login+"/", timeout)),
			len(exclude) > 0 && (excludes.RegexMatch(org.Login+"/someRepository", timeout) || excludes.RegexMatch(org.Login+"/", timeout)):

			continue
		}

		orgRepos, err := c.GetOrgRepos(ctx, org.Login)
		if err != nil {
			return nil, err
		}

		repos = append(repos, orgRepos...)

	}

	return repos, nil
}

// Get an organization.
func (c RESTClient) GetOrg(ctx context.Context, name string) (org *resources.Organization, err error) {
	err = c.DoWithContext(ctx, http.MethodGet, newRequestPath(orgEp.Format(map[string]any{"owner": name})).String(), nil, &org)
	return
}

// Get all repositories for given organization.
func (c RESTClient) GetOrgRepos(ctx context.Context, name string) ([]resources.Repository, error) {
	c.Progressbar.Describe("Retrieving repositories for GitHub organization: %s...", name)
	return getPaged[resources.Repository](c, orgReposEp.Format(map[string]any{"owner": name}), ctx)
}

// Get organizations.
func (c RESTClient) GetOrgs(ctx context.Context) ([]resources.Organization, error) {
	c.Progressbar.Describe("Retrieving GitHub organizations...")
	return getPaged[resources.Organization](c, orgsEp, ctx)
}

// Get rate limit information.
func (c RESTClient) GetRateLimit(ctx context.Context) (rate *resources.RateLimit, err error) {
	err = c.DoWithContext(ctx, http.MethodGet, newRequestPath(rateLimitEp).String(), nil, &rate)
	return
}

// Get all pull requests for given organization and repository.
func (c RESTClient) GetOrgRepoPulls(ctx context.Context, name, repo string, filter map[string]string) ([]resources.PullRequest, error) {
	c.Describe(fmt.Sprintf("Retrieving pull requests for GitHub repository: %s/%s...", name, repo))
	pulls, err := getPaged[resources.PullRequest](c, pullsEp.Format(map[string]any{"owner": name, "repo": repo}), ctx, func(params *requestPath) {
		params.
			Register("state", "open", "closed", "all").
			Register("sort", "created", "updated", "popularity", "long-running")
		for k, v := range filter {
			if v == "" {
				continue
			}
			params.Set(k, v)
		}
	})
	if err != nil {
		return nil, err
	}

	for i := range pulls {
		pulls[i].Owner = name
		pulls[i].Repository = repo
	}

	return pulls, nil
}

// Get GitHub user.
func (c RESTClient) GetUser(ctx context.Context) (user *resources.User, err error) {
	err = c.DoWithContext(ctx, http.MethodGet, newRequestPath(userEp).String(), nil, &user)
	return
}

// Get all repositories for given user.
func (c RESTClient) GetUserRepos(ctx context.Context) ([]resources.Repository, error) {
	c.Progressbar.Describe("Retrieving repositories for current user...")
	return getPaged[resources.Repository](c, userReposEp, ctx)
}

// get all organizations for given user.
func (c RESTClient) GetUserOrgs(ctx context.Context) ([]resources.Organization, error) {
	c.Progressbar.Describe("Retrieving GitHub organizations for current user...")
	return getPaged[resources.Organization](c, userOrgsEp, ctx)
}

// Implements RequestWithContext method.
func (c RESTClient) RequestWithContext(ctx context.Context, method string, path string, body io.Reader) (*http.Response, error) {
	return c.RESTClient.RequestWithContext(ctx, method, path, body)
}

// Create new REST API client.
// The rate limit of the API will be checked upfront.
func NewRESTClient(conf *configfile.Configuration, options ClientOptions) (*RESTClient, error) {
	loggerEntry.Debugf("Creating client with options: %+v", options)

	client, err := api.NewRESTClient(options)
	if err != nil {
		return nil, err
	}

	wrapClient := &RESTClient{
		RESTClient:    client,
		Configuration: conf,
		Progressbar:   util.NewProgressbar(-1),
	}

	rate, err := wrapClient.GetRateLimit(context.Background())
	if err != nil {
		return nil, err
	}

	defer CheckRateLimitAndExit(rate)
	return wrapClient, nil
}
