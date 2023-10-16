package configfile

import (
	"strings"

	auth "github.com/cli/go-gh/v2/pkg/auth"
	util "github.com/sarumaj/gh-gr/pkg/util"
)

func GetHosts() []string {
	hosts := auth.KnownHosts()
	if len(hosts) == 0 {
		host, _ := auth.DefaultHost()
		hosts = append(hosts, host)
	}

	return hosts
}

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

func IsRepoDir(path string, repos []Repository) bool {
	util.PathSanitize(&path)
	for _, r := range repos {
		util.PathSanitize(&r.Directory)
		if strings.HasPrefix(r.Directory+"/", path+"/") {
			return true
		}
	}

	return false
}

func newBinaryProgressbar() *util.Progressbar {
	c := util.Console()
	return util.NewProgressbar(
		-1,
		util.EnableColorCodes(c.ColorsEnabled()),
		util.SetWidth(10),
		util.ShowBytes(true),
		util.SetRenderBlankState(true),
		util.ClearOnFinish(),
		util.ShowCount(),
	)
}
