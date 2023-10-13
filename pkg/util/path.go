package util

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	supererrors "github.com/sarumaj/go-super/errors"
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

type popd string

func (p popd) Popd() {
	supererrors.Except(os.Chdir(string(p)))
}

func Chdir(path string) interface{ Popd() } {
	current := supererrors.ExceptFn(supererrors.W(os.Getwd()))

	supererrors.Except(os.Chdir(path))

	return popd(current)
}

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

func ListFilesByExtension(ext string) []string {
	var fileList []string
	supererrors.Except(filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ext {
			fileList = append(fileList, path)
		}

		return nil
	}))

	slices.Sort(fileList)
	return fileList
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	supererrors.Except(err, os.ErrNotExist)
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
