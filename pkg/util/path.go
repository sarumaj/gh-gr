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

// Regular expression sued to extract hostname from an URL.
var hostRegex = regexp.MustCompile(`(?:[^:]+://|[^/]*//)?(?:[^@]+@)?(?P<Hostname>[^/:]+).*`)

// Custom type with method to return to last working directory.
type popd string

// Return to last saved directory.
func (p popd) Popd() {
	supererrors.Except(os.Chdir(string(p)))
}

// Change current working directory.
func Chdir(path string) interface{ Popd() } {
	current := supererrors.ExceptFn(supererrors.W(os.Getwd()))

	supererrors.Except(os.Chdir(path))

	return popd(current)
}

// Get path of the executable that started current gr process.
func GetExecutablePath() string {
	executablePath, err := os.Executable()
	if err != nil {
		return "unknown"
	}

	evaluatedPath, err := filepath.EvalSymlinks(executablePath)
	if err != nil {
		return executablePath
	}

	return evaluatedPath
}

// Extract hostname from path.
func GetHostnameFromPath(path string) string {
	return hostRegex.ReplaceAllString(path, "$Hostname")
}

// Format integer as byte size in kMGTPE[B] unit.
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

// List files in current working directory matching given extension (extension has to start with a dot).
func ListFilesByExtension(ext string, depth int) []string {
	var fileList []string
	supererrors.Except(filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if depth > 0 && strings.Count(filepath.ToSlash(path), "/") > depth {
			return nil
		}

		if !info.IsDir() && filepath.Ext(path) == ext {
			fileList = append(fileList, path)
		}

		return nil
	}))

	slices.Sort(fileList)
	return fileList
}

// Check if path exists.
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

// Format path to UNIX style path, strip volume name if default.
func PathSanitize(paths ...*string) {
	for _, path := range paths {
		if path == nil {
			continue
		}

		*path = filepath.Clean(*path)

		if volume := filepath.VolumeName(*path); strings.EqualFold(volume, "C:") {
			*path = strings.TrimPrefix(*path, volume)
		}

		*path = filepath.ToSlash(*path)
		*path = strings.TrimSuffix(*path, "/")
	}
}

// StripPathPrefix strips prefix from path given a number of parent directories to keep.
func StripPathPrefix(path string, keepParents uint) string {
	if path == "" {
		return filepath.FromSlash("")
	}

	if p, err := filepath.Abs(path); err == nil {
		path = p
	}

	if length := len(filepath.SplitList(path)); length > int(keepParents) {
		keepParents = uint(length)
	}

	parts := []string{filepath.Base(path)}
	for i := uint(0); i < keepParents; i++ {
		path = filepath.Dir(path)
		parts = append([]string{filepath.Base(path)}, parts...)
	}

	return filepath.Join(parts...)
}
