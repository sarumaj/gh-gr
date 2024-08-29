package util

import (
	"path/filepath"
	"time"

	regexp2 "github.com/dlclark/regexp2"
)

// PatternList is a list of regular expressions.
type PatternList []string

// GlobMatch checks if target matches any of the globs in the list.
func (l PatternList) GlobMatch(target string) bool {
	for _, pattern := range l {
		ok, err := filepath.Match(pattern, target)
		if err != nil {
			continue
		}

		if ok {
			return true
		}
	}

	return false
}

// GlobMatchAny checks if any of the targets match any of the globs in the list.
func (l PatternList) GlobMatchAny(targets ...string) bool {
	for _, target := range targets {
		if l.GlobMatch(target) {
			return true
		}
	}

	return false
}

// RegexMatch checks if target matches any of the regular expressions in the list.
func (l PatternList) RegexMatch(target string, timeout time.Duration) bool {
	for _, pattern := range l {
		re, err := regexp2.Compile(pattern, regexp2.RE2)
		if err != nil {
			continue
		}

		if timeout > 0 {
			re.MatchTimeout = timeout
		}

		if match, err := re.MatchString(target); err == nil && match {
			return true
		}
	}

	return false
}

// RegexMatchAny checks if any of the targets match any of the regular expressions in the list.
func (l PatternList) RegexMatchAny(timeout time.Duration, targets ...string) bool {
	for _, target := range targets {
		if l.RegexMatch(target, timeout) {
			return true
		}
	}

	return false
}
