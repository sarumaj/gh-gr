package restclient

import (
	"errors"
	"fmt"
	"time"

	"github.com/sarumaj/gh-pr/pkg/restclient/resources"
)

func CheckRateLimit(r *resources.RateLimit) error {
	if r.Resources.Core.Remaining == 0 {
		resetTime := time.Unix(r.Resources.Core.Reset, 0)
		msg := fmt.Sprintf(
			"gh REST API quota have been exhausted. Please, wait until %s (%s remaining...)",
			resetTime, time.Until(resetTime),
		)
		return errors.New(msg)
	}

	return nil
}
