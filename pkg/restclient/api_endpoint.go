package restclient

import (
	"fmt"
	"strings"
)

const (
	orgEp       = apiEndpoint("orgs/{owner}")
	orgReposEp  = apiEndpoint("orgs/{owner}/repos")
	orgsEp      = apiEndpoint("organizations")
	pullEp      = apiEndpoint("repos/{owner}/{repo}/pulls/{number}")
	pullsEp     = apiEndpoint("repos/{owner}/{repo}/pulls")
	rateLimitEp = apiEndpoint("rate_limit")
	userEp      = apiEndpoint("user")
	userOrgsEp  = apiEndpoint("user/orgs")
	userReposEp = apiEndpoint("user/repos")
)

// Helper for storing API endpoints.
type apiEndpoint string

// Replace the {KEY} placeholders with values from a map.
func (s apiEndpoint) Format(params map[string]any) apiEndpoint {
	var replacements []string
	for k, v := range params {
		replacements = append(replacements, fmt.Sprintf("{%s}", k), fmt.Sprintf("%v", v))
	}

	return apiEndpoint(strings.NewReplacer(replacements...).Replace(string(s)))
}
