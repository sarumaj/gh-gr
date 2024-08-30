package restclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
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
	rateMutex sync.Mutex
	rateReset time.Time
	retry     bool
}

// Close a pull request.
func (c *RESTClient) ClosePullRequest(ctx context.Context, owner, repo string, number int) error {
	c.Describe(fmt.Sprintf("Closing pull request %d for GitHub repository: %s/%s...", number, owner, repo))
	return c.DoWithContext(ctx, http.MethodPatch,
		newRequestPath(pullEp.Format(map[string]any{"owner": owner, "repo": repo, "number": number})).String(),
		strings.NewReader(`{"state":"closed"}`), nil)
}

// Overwrites DoWithContext method.
func (c *RESTClient) DoWithContext(ctx context.Context, method string, path string, body io.Reader, response any) error {
	resp, err := c.RequestWithContext(ctx, method, path, body)
	if err != nil {
		return err
	}

	return json.NewDecoder(resp.Body).Decode(response)
}

// Get all repositories for given user and the organizations he belongs to.
func (c *RESTClient) GetAllUserRepos(ctx context.Context, include, exclude []string) ([]resources.Repository, error) {
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
func (c *RESTClient) GetOrg(ctx context.Context, name string) (org *resources.Organization, err error) {
	err = c.DoWithContext(ctx, http.MethodGet, newRequestPath(orgEp.Format(map[string]any{"owner": name})).String(), nil, &org)
	return
}

// Get all repositories for given organization.
func (c *RESTClient) GetOrgRepos(ctx context.Context, name string) ([]resources.Repository, error) {
	c.Progressbar.Describe("Retrieving repositories for GitHub organization: %s...", name)
	return getPaged[resources.Repository, []resources.Repository](c, orgReposEp.Format(map[string]any{"owner": name}), ctx)
}

// Get organizations.
func (c *RESTClient) GetOrgs(ctx context.Context) ([]resources.Organization, error) {
	c.Progressbar.Describe("Retrieving GitHub organizations...")
	return getPaged[resources.Organization, []resources.Organization](c, orgsEp, ctx)
}

// Get rate limit information.
func (c *RESTClient) GetRateLimit(ctx context.Context) (rate *resources.RateLimit, headers http.Header, err error) {
	resp, err := c.RequestWithContext(ctx, http.MethodGet, newRequestPath(rateLimitEp).String(), nil)
	if err != nil {
		return nil, nil, err
	}

	if err := json.NewDecoder(resp.Body).Decode(&rate); err != nil {
		return nil, nil, err
	}

	return rate, resp.Header, nil
}

// Get all pull requests for given organization and repository.
func (c *RESTClient) GetOrgRepoPulls(ctx context.Context, name, repo string, filter map[string]string) (out []resources.PullRequest, err error) {
	c.Describe(fmt.Sprintf("Retrieving pull requests for GitHub repository: %s/%s...", name, repo))
	pulls, err := getPaged[resources.PullRequest, []resources.PullRequest](c, pullsEp.Format(map[string]any{"owner": name, "repo": repo}), ctx, func(params *requestPath) {
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

	for i, pull := range pulls {
		if err := c.DoWithContext(ctx, http.MethodGet, pull.URL, nil, &pulls[i]); err != nil {
			return nil, err
		}
		pulls[i].Repository = name + "/" + repo
	}

	return pulls, nil
}

// Get GitHub user.
func (c *RESTClient) GetUser(ctx context.Context) (user *resources.User, err error) {
	err = c.DoWithContext(ctx, http.MethodGet, newRequestPath(userEp).String(), nil, &user)
	return
}

// Get all repositories for given user.
func (c *RESTClient) GetUserRepos(ctx context.Context) ([]resources.Repository, error) {
	c.Progressbar.Describe("Retrieving repositories for current user...")
	return getPaged[resources.Repository, []resources.Repository](c, userReposEp, ctx)
}

// get all organizations for given user.
func (c *RESTClient) GetUserOrgs(ctx context.Context) ([]resources.Organization, error) {
	c.Progressbar.Describe("Retrieving GitHub organizations for current user...")
	return getPaged[resources.Organization, []resources.Organization](c, userOrgsEp, ctx)
}

// Reopen a pull request.
func (c *RESTClient) ReopenPullRequest(ctx context.Context, owner, repo string, number int) error {
	c.Describe(fmt.Sprintf("Reopening pull request %d for GitHub repository: %s/%s...", number, owner, repo))
	return c.DoWithContext(ctx, http.MethodPatch,
		newRequestPath(pullEp.Format(map[string]any{"owner": owner, "repo": repo, "number": number})).String(),
		strings.NewReader(`{"state":"open"}`), nil)
}

// Overwrites RequestWithContext method.
func (c *RESTClient) RequestWithContext(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	c.rateMutex.Lock()
	if timeUntilReset := time.Until(c.rateReset); timeUntilReset > 0 && c.retry {
		if timeUntilReset > 0 {
			time.Sleep(timeUntilReset)
		}
	}
	c.rateMutex.Unlock()

	resp, err := c.RESTClient.RequestWithContext(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	remaining, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
	if err != nil {
		return nil, err
	}

	reset, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Reset"))
	if err != nil {
		return nil, err
	}

	c.rateMutex.Lock()
	if remaining == 0 && c.retry {
		c.rateReset = time.Unix(int64(reset), 0).Add(time.Second)
		c.rateMutex.Unlock()
		c.Describe(fmt.Sprintf("Rate limit exceeded, waiting for %s...", time.Until(c.rateReset)))
		return c.RequestWithContext(ctx, method, path, body)
	}

	c.rateMutex.Unlock()
	return resp, nil
}

// Search for pull requests in a repository.
func (c *RESTClient) SearchOrgRepoPulls(ctx context.Context, name, repo string, filter string) (out []resources.PullRequest, err error) {
	c.Describe(fmt.Sprintf("Retrieving pull requests for GitHub repository: %s/%s...", name, repo))
	searchQuery := fmt.Sprintf("is:pr repo:%s/%s", name, repo)
	if filter != "" {
		searchQuery += " " + filter
	}

	searchResults, err := getPaged[resources.PullRequest, resources.SearchResult[resources.PullRequest]](c, searchIssuesEp, ctx, func(rp *requestPath) {
		rp.Set("q", searchQuery)
	})
	if err != nil {
		return nil, err
	}

	for _, item := range searchResults.Items {
		var pr resources.PullRequest
		if err := c.DoWithContext(ctx, http.MethodGet, item.URL, nil, &pr); err != nil {
			return nil, err
		}

		pr.Repository = name + "/" + repo
		out = append(out, pr)
	}

	return out, nil
}

// Create new REST API client.
// The rate limit of the API will be checked upfront.
func NewRESTClient(conf *configfile.Configuration, options ClientOptions, retry bool) (*RESTClient, error) {
	loggerEntry.Debugf("Creating client with options: %+v", options)

	client, err := api.NewRESTClient(options)
	if err != nil {
		return nil, err
	}

	wrapClient := &RESTClient{
		RESTClient:    client,
		Configuration: conf,
		Progressbar:   util.NewProgressbar(-1),
		retry:         retry,
	}

	rate, _, err := wrapClient.GetRateLimit(context.Background())
	if err != nil {
		return nil, err
	}

	defer CheckRateLimitAndExit(rate)
	return wrapClient, nil
}
