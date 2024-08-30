package configfile

import (
	"net/url"
	"path/filepath"
	"strings"

	auth "github.com/cli/go-gh/v2/pkg/auth"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
)

// Retrieve all configured hosts from GitHub CLI.
func GetHosts() []string {
	hosts := auth.KnownHosts()
	if len(hosts) == 0 {
		host, _ := auth.DefaultHost()
		hosts = append(hosts, host)
	}

	return hosts
}

// GetRepositorySlugFromURL extracts repository slug from URL.
func GetRepositorySlugFromURL(repo Repository) string {
	parsed, _ := url.Parse(repo.URL)
	return strings.TrimSuffix(strings.TrimPrefix(parsed.Path, "/"), filepath.Ext(parsed.Path))
}

// Retrieve all authentication tokens for each host from GitHub CLI.
func GetTokens() map[string]string {
	tokens := make(map[string]string)

	for _, host := range GetHosts() {
		host = util.GetHostnameFromPath(host)
		loggerEntry.Debugf("Retrieving token for host: %s", host)

		token, _ := auth.TokenForHost(host)
		loggerEntry.Debugf("Retrieved token: %t", len(token) > 0)

		tokens[host] = token
	}

	return tokens
}

// Check if existing directory is enlisted as GitHub repository.
func isRepoDir(path string, repos []Repository) bool {
	util.PathSanitize(&path)
	for _, r := range repos {
		util.PathSanitize(&r.Directory)
		if strings.HasPrefix(r.Directory+"/", path+"/") {
			return true
		}
	}

	return false
}

// Create progressbar for binary data stream (unknown length).
func newBinaryProgressbar() *util.Progressbar {
	c := util.Console()
	width, _, _ := c.Size()
	width = min(width/10, 20)

	return util.NewProgressbar(
		-1,
		util.EnableColorCodes(c.ColorsEnabled()),
		util.SetWidth(width),
		util.ShowBytes(true),
		util.SetRenderBlankState(true),
		util.ClearOnFinish(),
		util.ShowCount(),
	)
}
