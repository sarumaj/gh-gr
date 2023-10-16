package restclient

import (
	"time"

	color "github.com/fatih/color"
	resources "github.com/sarumaj/gh-gr/pkg/restclient/resources"
	util "github.com/sarumaj/gh-gr/pkg/util"
)

// Message emitted, when API limit exceeded.
const RateLimitExceeded = "GitHub REST API quotas have been exhausted. Please, wait until %s (%s remaining...)"

// Checks whether the quota of the core API are exceeded.
func CheckRateLimitAndExit(r *resources.RateLimit) {
	if r.Resources.Core.Remaining > 0 {
		return
	}

	c := util.Console()
	resetTime := time.Unix(r.Resources.Core.Reset, 0)
	util.PrintlnAndExit(c.CheckColors(color.RedString, RateLimitExceeded, resetTime, time.Until(resetTime)))
}
