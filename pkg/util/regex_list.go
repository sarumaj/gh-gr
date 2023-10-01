package util

import "regexp"

type RegexList []string

func (l RegexList) Match(target string) bool {
	for _, regex := range l {
		re, err := regexp.Compile(regex)
		if err == nil && re.MatchString(target) {
			return true
		}
	}

	return false
}
