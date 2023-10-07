package restclient

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/sarumaj/gh-gr/pkg/restclient/resources"
	"github.com/sarumaj/gh-gr/pkg/util"
)

const RateLimitExceeded = "GitHub REST API quotas have been exhausted. Please, wait until %s (%s remaining...)"

func CheckRateLimit(r *resources.RateLimit) {
	if r.Resources.Core.Remaining == 0 {
		resetTime := time.Unix(r.Resources.Core.Reset, 0)
		fmt.Fprintln(util.Stderr(), util.CheckColors(color.RedString, RateLimitExceeded, resetTime, time.Until(resetTime)))

		os.Exit(1)
	}
}
