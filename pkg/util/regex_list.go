package util

import (
	"time"

	"github.com/dlclark/regexp2"
)

type RegexList []string

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
