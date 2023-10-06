package util

import (
	"os"
	"regexp"
)

var hostRegex = regexp.MustCompile(`(?:[^:]+://|[^/]*//)?(?P<Hostname>[^/:]+).*`)

func GetHostnameFromPath(path string) string {
	return hostRegex.ReplaceAllString(path, "$Hostname")
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	FatalIfError(err)
	return false
}
