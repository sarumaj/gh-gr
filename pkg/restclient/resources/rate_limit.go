package resources

type RateLimit struct {
	Resources RateResources `json:"resources"`
	Rate      Rate          `json:"rate"` // Deprecated: Resources.Core should be used instead.
}
