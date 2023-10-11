package restclient

import (
	"time"

	color "github.com/fatih/color"
	resources "github.com/sarumaj/gh-gr/pkg/restclient/resources"
	util "github.com/sarumaj/gh-gr/pkg/util"
)

const RateLimitExceeded = "GitHub REST API quotas have been exhausted. Please, wait until %s (%s remaining...)"

func CheckRateLimitAndExit(r *resources.RateLimit) {
	if r.Resources.Core.Remaining > 0 {
		return
	}

	resetTime := time.Unix(r.Resources.Core.Reset, 0)
	util.PrintlnAndExit(util.CheckColors(color.RedString, RateLimitExceeded, resetTime, time.Until(resetTime)))
}
