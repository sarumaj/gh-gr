package util

import (
	"time"

	"github.com/dlclark/regexp2"
)

// Custom type to implement regex matcher for any of.
type RegexList []string

// Match at least once for given target.
func (l RegexList) Match(target string, timeout time.Duration) bool {
	for _, regex := range l {
		re, err := regexp2.Compile(regex, regexp2.RE2)
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
