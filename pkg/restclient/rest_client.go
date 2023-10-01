package restclient

import (
	"context"
	"io"
	"net/http"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/sarumaj/gh-pr/pkg/configfile"
	"github.com/sarumaj/gh-pr/pkg/restclient/resources"
)

type ClientOptions = api.ClientOptions

type RESTClient struct {
	*api.RESTClient
	*configfile.Configuration
}

func (c RESTClient) do(ctx context.Context, method, path string, body io.Reader, response any) error {
	return c.DoWithContext(ctx, method, path, body, response)
}

func (c RESTClient) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	return c.RequestWithContext(ctx, method, path, body)
}

func (c RESTClient) GetOrg(ctx context.Context, name string) (org *resources.Organization, err error) {
	err = c.do(ctx, http.MethodGet, newRequestPath(orgEp.Format(map[string]any{"org": name})).String(), nil, &org)
	return
}

func (c RESTClient) GetOrgRepos(ctx context.Context, name string) ([]resources.Repository, error) {
	return getPaged[resources.Repository](c, orgReposEp.Format(map[string]any{"org": name}), ctx)
}

func (c RESTClient) GetOrgs(ctx context.Context) ([]resources.Organization, error) {
	return getPaged[resources.Organization](c, orgsEp, ctx)
}

func (c RESTClient) GetRateLimit(ctx context.Context) (rate *resources.RateLimit, err error) {
	err = c.do(ctx, http.MethodGet, newRequestPath(rateLimitEp).String(), nil, &rate)
	return
}

func (c RESTClient) GetUser(ctx context.Context) (user *resources.User, err error) {
	err = c.do(ctx, http.MethodGet, newRequestPath(userEp).String(), nil, &user)
	return
}

func (c RESTClient) GetUserRepos(ctx context.Context) ([]resources.Repository, error) {
	return getPaged[resources.Repository](c, userReposEp, ctx)
}

func (c RESTClient) GetUserOrgs(ctx context.Context) ([]resources.Organization, error) {
	return getPaged[resources.Organization](c, userOrgsEp, ctx)
}

func NewRESTClient(conf *configfile.Configuration, options ClientOptions) (*RESTClient, error) {
	client, err := api.NewRESTClient(options)
	if err != nil {
		return nil, err
	}

	return &RESTClient{RESTClient: client, Configuration: conf}, nil
}
