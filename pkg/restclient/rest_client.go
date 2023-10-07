package restclient

import (
	"context"
	"net/http"

	api "github.com/cli/go-gh/v2/pkg/api"
	configfile "github.com/sarumaj/gh-gr/pkg/configfile"
	resources "github.com/sarumaj/gh-gr/pkg/restclient/resources"
	util "github.com/sarumaj/gh-gr/pkg/util"
)

var loggerEntry = util.Logger.WithField("mod", "restclient")

type ClientOptions = api.ClientOptions

type RESTClient struct {
	*api.RESTClient
	*configfile.Configuration
	*util.Progressbar
}

func (c RESTClient) GetOrg(ctx context.Context, name string) (org *resources.Organization, err error) {
	err = c.DoWithContext(ctx, http.MethodGet, newRequestPath(orgEp.Format(map[string]any{"org": name})).String(), nil, &org)
	return
}

func (c RESTClient) GetOrgRepos(ctx context.Context, name string) ([]resources.Repository, error) {
	c.Progressbar.Describe("Retrieving repositories for GitHub organization: %s...", name)
	return getPaged[resources.Repository](c, orgReposEp.Format(map[string]any{"org": name}), ctx)
}

func (c RESTClient) GetOrgs(ctx context.Context) ([]resources.Organization, error) {
	c.Progressbar.Describe("Retrieving GitHub organizations...")
	return getPaged[resources.Organization](c, orgsEp, ctx)
}

func (c RESTClient) GetRateLimit(ctx context.Context) (rate *resources.RateLimit, err error) {
	err = c.DoWithContext(ctx, http.MethodGet, newRequestPath(rateLimitEp).String(), nil, &rate)
	return
}

func (c RESTClient) GetUser(ctx context.Context) (user *resources.User, err error) {
	err = c.DoWithContext(ctx, http.MethodGet, newRequestPath(userEp).String(), nil, &user)
	return
}

func (c RESTClient) GetUserRepos(ctx context.Context) ([]resources.Repository, error) {
	c.Progressbar.Describe("Retrieving repositories for current user...")
	return getPaged[resources.Repository](c, userReposEp, ctx)
}

func (c RESTClient) GetUserOrgs(ctx context.Context) ([]resources.Organization, error) {
	c.Progressbar.Describe("Retrieving GitHub organizations for current user...")
	return getPaged[resources.Organization](c, userOrgsEp, ctx)
}

func NewRESTClient(conf *configfile.Configuration, options ClientOptions) (*RESTClient, error) {
	loggerEntry.Debugf("Creating client with options: %+v", options)

	client, err := api.NewRESTClient(options)
	if err != nil {
		return nil, err
	}

	return &RESTClient{
		RESTClient:    client,
		Configuration: conf,
		Progressbar:   util.NewProgressbar(-1),
	}, nil
}
