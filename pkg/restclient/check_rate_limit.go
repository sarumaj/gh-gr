package restclient

import (
	"time"

	color "github.com/fatih/color"
	resources "github.com/sarumaj/gh-gr/v2/pkg/restclient/resources"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
)

// Message emitted, when API limit exceeded.
const RateLimitExceeded = "GitHub REST API quotas have been exhausted. Please, wait until %s (%s remaining...)"

// Checks whether the quota of the core API are exceeded.
func CheckRateLimitAndExit(r *resources.RateLimit) {
	if r.Resources.Core.Remaining > 0 && r.Resources.Search.Remaining > 0 {
		return
	}

	check := func(rate resources.Rate) string {
		if rate.Remaining > 0 {
			return ""
		}

		c := util.Console()
		resetTime := time.Unix(rate.Reset, 0)

		return c.CheckColors(color.RedString, RateLimitExceeded, resetTime, time.Until(resetTime))
	}

	if msg := check(r.Resources.Core); msg != "" {
		util.PrintlnAndExit("%s", msg)
	}

	if msg := check(r.Resources.Search); msg != "" {
		util.PrintlnAndExit("%s", msg)
	}
}
