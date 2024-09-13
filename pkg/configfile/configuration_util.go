package configfile

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	terminal "github.com/AlecAivazis/survey/v2/terminal"
	auth "github.com/cli/go-gh/v2/pkg/auth"
	browser "github.com/cli/go-gh/v2/pkg/browser"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	supererrors "github.com/sarumaj/go-super/errors"
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

// Open links in browser.
func OpenLins(links []string) {
	c := util.Console()
	client := browser.New("", c.Stdout(), c.Stderr())
	for len(links) > 0 {
		choice := supererrors.ExceptFn(supererrors.W(
			prompt.Select(
				"Select a link to open:",
				links[0],
				links,
			),
		), terminal.InterruptErr)

		if supererrors.LastErrorWas(terminal.InterruptErr) || choice >= len(links) {
			os.Exit(0)
		}

		supererrors.Except(client.Browse(links[choice]))
		links = append(links[:choice], links[choice+1:]...)[: len(links)-1 : len(links)-1]
	}

	if len(links) == 0 {
		os.Exit(0)
	}
}
