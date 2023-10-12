package util

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	KB = 1 << (10 * (iota + 1))
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

var hostRegex = regexp.MustCompile(`(?:[^:]+://|[^/]*//)?(?:[^@]+@)?(?P<Hostname>[^/:]+).*`)

func IntToSizeBytes(s int, unit int64, precision int) string {
	b := int64(s)
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf(
		fmt.Sprintf("%%.%df %%cB", precision),
		float64(b)/float64(div),
		"kMGTPE"[exp],
	)
}

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
