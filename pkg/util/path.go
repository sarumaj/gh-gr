package util

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var hostRegex = regexp.MustCompile(`(?:[^:]+://|[^/]*//)?(?P<Hostname>[^/:]+).*`)

func GetHostnameFromPath(path string) string {
	return hostRegex.ReplaceAllString(path, "$Hostname")
}

func MoveToPath(path string) (back func()) {
	current, err := os.Getwd()
	FatalIfError(err)

	FatalIfError(os.Chdir(path))

	return func() { FatalIfError(os.Chdir(current)) }
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

func PathSanitize(paths ...*string) {
	for _, path := range paths {
		if path == nil {
			continue
		}

		*path = filepath.Clean(*path)

		if volume := filepath.VolumeName(*path); volume == "C:" || volume == "c:" {
			*path = strings.Replace(*path, volume, "", 1)
		}

		*path = filepath.ToSlash(*path)
		*path = strings.TrimSuffix(*path, "/")
	}
}
