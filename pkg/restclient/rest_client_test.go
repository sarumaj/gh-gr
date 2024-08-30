package restclient

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/sarumaj/gh-gr/v2/pkg/configfile"
)

func setupTestServer(tb testing.TB) *httptest.Server {
	tb.Helper()

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("test", r.URL.Path+".json"))
	}))
	server.TLS = &tls.Config{}

	server.StartTLS()
	tb.Cleanup(server.Close)

	return server
}

func setupTestClient(tb testing.TB, server *httptest.Server) *RESTClient {
	tb.Helper()

	parsed, err := url.Parse(server.URL)
	if err != nil {
		tb.Fatalf("Failed to parse server URL: %v", err)
	}

	client, err := NewRESTClient(
		&configfile.Configuration{Concurrency: 16, Timeout: time.Hour},
		ClientOptions{
			AuthToken: "1234",
			Host:      parsed.Host,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		},
		true,
	)
	if err != nil {
		tb.Fatalf("Failed to create REST client: %v", err)
	}

	tb.Cleanup(func() { _ = client.Close() })
	return client
}

func TestRESTClient(t *testing.T) {
	server := setupTestServer(t)
	client := setupTestClient(t, server)

	t.Run("ClosePullRequest", func(t *testing.T) {
		// spellchecker:ignore octocat
		if err := client.ClosePullRequest(context.TODO(), "octocat", "Hello-World", 1347); err != nil {
			t.Fatalf("Failed to close pull request: %v", err)
		}
	})

	t.Run("GetAllUserRepos", func(t *testing.T) {
		if repos, err := client.GetAllUserRepos(context.TODO(), nil, nil); err != nil {
			t.Fatalf("Failed to get user repositories: %v", err)
		} else if len(repos) == 0 {
			t.Fatalf("Failed to get user repositories: no repositories found")
		}
	})

	t.Run("GetOrgRepos", func(t *testing.T) {
		if repos, err := client.GetOrgRepos(context.TODO(), "github"); err != nil {
			t.Fatalf("Failed to get org repos: %v", err)
		} else if len(repos) == 0 {
			t.Fatalf("Failed to get org repos: no repositories found")
		}
	})

	t.Run("GetUserOrgs", func(t *testing.T) {
		if orgs, err := client.GetUserOrgs(context.TODO()); err != nil {
			t.Fatalf("Failed to get user orgs: %v", err)
		} else if len(orgs) == 0 {
			t.Fatalf("Failed to get user orgs: no organizations found")
		}
	})

	t.Run("GetUserRepos", func(t *testing.T) {
		if repos, err := client.GetUserRepos(context.TODO()); err != nil {
			t.Fatalf("Failed to get user repos: %v", err)
		} else if len(repos) == 0 {
			t.Fatalf("Failed to get user repos: no repositories found")
		}
	})

	t.Run("GetOrg", func(t *testing.T) {
		if org, err := client.GetOrg(context.TODO(), "github"); err != nil {
			t.Fatalf("Failed to get org: %v", err)
		} else if org == nil {
			t.Fatalf("Failed to get org: no organization found")
		}
	})

	t.Run("GetOrgs", func(t *testing.T) {
		if orgs, err := client.GetOrgs(context.TODO()); err != nil {
			t.Fatalf("Failed to get orgs: %v", err)
		} else if len(orgs) == 0 {
			t.Fatalf("Failed to get orgs: no organizations found")
		}
	})

	t.Run("GetRateLimit", func(t *testing.T) {
		if rate, _, err := client.GetRateLimit(context.TODO()); err != nil {
			t.Fatalf("Failed to get rate limit: %v", err)
		} else if rate == nil {
			t.Fatalf("Failed to get rate limit: no rate limit found")
		}
	})

	t.Run("GetOrgRepoPulls", func(t *testing.T) {
		if pulls, err := client.GetOrgRepoPulls(context.TODO(), "octocat", "Hello-World", map[string]string{}); err != nil {
			t.Fatalf("Failed to get org repo pulls: %v", err)
		} else if len(pulls) == 0 {
			t.Fatalf("Failed to get org repo pulls: no pull requests found")
		}
	})

	t.Run("GetUser", func(t *testing.T) {
		if user, err := client.GetUser(context.TODO()); err != nil {
			t.Fatalf("Failed to get user: %v", err)
		} else if user == nil {
			t.Fatalf("Failed to get user: no user found")
		}
	})

	t.Run("GetUserRepos", func(t *testing.T) {
		if repos, err := client.GetUserRepos(context.TODO()); err != nil {
			t.Fatalf("Failed to get user repos: %v", err)
		} else if len(repos) == 0 {
			t.Fatalf("Failed to get user repos: no repositories found")
		}
	})

	t.Run("GetUserOrgs", func(t *testing.T) {
		if orgs, err := client.GetUserOrgs(context.TODO()); err != nil {
			t.Fatalf("Failed to get user orgs: %v", err)
		} else if len(orgs) == 0 {
			t.Fatalf("Failed to get user orgs: no organizations found")
		}
	})

	t.Run("ReopenPullRequest", func(t *testing.T) {
		// spellchecker:ignore octocat
		if err := client.ReopenPullRequest(context.TODO(), "octocat", "Hello-World", 1347); err != nil {
			t.Fatalf("Failed to reopen pull request: %v", err)
		}
	})

	t.Run("SearchOrgRepoPulls", func(t *testing.T) {
		if pulls, err := client.SearchOrgRepoPulls(context.TODO(), "octocat", "Hello-World", ""); err != nil {
			t.Fatalf("Failed to search org repo pulls: %v", err)
		} else if len(pulls) == 0 {
			t.Fatalf("Failed to search org repo pulls: no pull requests found")
		}
	})
}
