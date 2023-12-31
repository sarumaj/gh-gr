package restclient

import (
	"fmt"
	"strings"
)

const (
	orgEp       = apiEndpoint("orgs/{org}")
	orgReposEp  = apiEndpoint("orgs/{org}/repos")
	orgsEp      = apiEndpoint("organizations")
	rateLimitEp = apiEndpoint("rate_limit")
	userEp      = apiEndpoint("user")
	userOrgsEp  = apiEndpoint("user/orgs")
	userReposEp = apiEndpoint("user/repos")
)

// Helper for storing API endpoints.
type apiEndpoint string

// Replace the {KEY} placeholders with values from a map.
func (s apiEndpoint) Format(params map[string]any) apiEndpoint {
	o := s
	for k, v := range params {
		o = apiEndpoint(strings.ReplaceAll(string(s), "{"+k+"}", fmt.Sprint(v)))
	}

	return o
}
