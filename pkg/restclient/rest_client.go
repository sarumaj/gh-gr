package restclient

import (
	"context"
	"io"
	"net/http"

	api "github.com/cli/go-gh/v2/pkg/api"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	resources "github.com/sarumaj/gh-gr/pkg/restclient/resources"
	util "github.com/sarumaj/gh-gr/pkg/util"
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

// Implements DoWithContext method.
func (c RESTClient) DoWithContext(ctx context.Context, method string, path string, body io.Reader, response any) error {
	// reserved for future implementations
	return c.RESTClient.DoWithContext(ctx, method, path, body, response)
}

// Get all repositories for given user and the organizations he belongs to.
func (c RESTClient) GetAllUserRepos(ctx context.Context) ([]resources.Repository, error) {
	repos, err := c.GetUserRepos(ctx)
	if err != nil {
		return nil, err
	}

	orgs, err := c.GetUserOrgs(ctx)
	if err != nil {
		return nil, err
	}

	for _, org := range orgs {
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
	err = c.DoWithContext(ctx, http.MethodGet, newRequestPath(orgEp.Format(map[string]any{"org": name})).String(), nil, &org)
	return
}

// Get all repositories for given organization.
func (c RESTClient) GetOrgRepos(ctx context.Context, name string) ([]resources.Repository, error) {
	c.Progressbar.Describe("Retrieving repositories for GitHub organization: %s...", name)
	return getPaged[resources.Repository](c, orgReposEp.Format(map[string]any{"org": name}), ctx)
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
	// reserved for future implementations
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
