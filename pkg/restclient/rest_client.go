package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/sarumaj/gh-pr/pkg/config"
	"github.com/sarumaj/gh-pr/pkg/resources"
	"github.com/sirupsen/logrus"
)

type ClientOptions = api.ClientOptions

type RESTClient struct {
	*api.RESTClient
	*logrus.Entry
}

func (c RESTClient) do(ctx context.Context, method, path string, body io.Reader, response any) error {
	c.Debugf("Sending REST call: %s %s to retrieve %T", method, path, response)
	return c.DoWithContext(ctx, method, path, body, response)
}

func (c RESTClient) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	c.Debugf("Sending REST call: %s %s", method, path)
	return c.RequestWithContext(ctx, method, path, body)
}

func (c RESTClient) GetOrg(ctx context.Context, name string) (org *resources.Organization, err error) {
	err = c.do(ctx, http.MethodGet, newRequestPath(orgEp.Format(map[string]any{"org": name})).String(), nil, &org)
	return
}

func (c RESTClient) GetOrgRepos(ctx context.Context, name string) (resources.Repositories, error) {
	return getPaged[resources.Repository](c, orgReposEp.Format(map[string]any{"org": name}), ctx)
}

func (c RESTClient) GetOrgs(ctx context.Context) (resources.Organizations, error) {
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

func (c RESTClient) GetUserRepos(ctx context.Context) (resources.Repositories, error) {
	return getPaged[resources.Repository](c, userReposEp, ctx)
}

func (c RESTClient) GetUserOrgs(ctx context.Context) (resources.Organizations, error) {
	return getPaged[resources.Organization](c, userOrgsEp, ctx)
}

func NewRESTClient(options ClientOptions) (*RESTClient, error) {
	client, err := api.NewRESTClient(options)
	if err != nil {
		return nil, err
	}

	logger := config.Logger()
	return &RESTClient{
		RESTClient: client,
		Entry: logger.WithFields(logrus.Fields{
			"RestClientID": fmt.Sprintf("%p", client),
		}),
	}, nil
}
