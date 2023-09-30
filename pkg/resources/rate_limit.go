package resources

import "time"

type RateLimit struct {
	Resources RateResources `json:"resources"`
	Rate      Rate          `json:"rate"` // Deprecated: Resources.Core should be used instead.
}

func (r RateLimit) GetResetTime() time.Time {
	return time.Unix(r.Resources.Core.Reset, 0)
}

func (r RateLimit) IsExhausted() bool {
	return r.Resources.Core.Remaining == 0
}
