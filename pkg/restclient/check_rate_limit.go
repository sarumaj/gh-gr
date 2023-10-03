package restclient

import (
	"errors"
	"fmt"
	"time"

	"github.com/sarumaj/gh-gr/pkg/restclient/resources"
)

func CheckRateLimit(r *resources.RateLimit) error {
	if r.Resources.Core.Remaining == 0 {
		resetTime := time.Unix(r.Resources.Core.Reset, 0)
		msg := fmt.Sprintf(
			"GitHub REST API quotas have been exhausted. Please, wait until %s (%s remaining...)",
			resetTime, time.Until(resetTime),
		)
		return errors.New(msg)
	}

	return nil
}
