package restclient

import (
	"net/url"
)

// Helper for building request paths.
type requestPath struct {
	Endpoint    apiEndpoint
	Querystring url.Values
}

// Add querystring params.
func (r *requestPath) Add(key string, values ...string) *requestPath {
	for _, value := range values {
		r.Querystring.Add(key, value)
	}

	return r
}

// Delete querystring param.
func (r *requestPath) Del(key string) *requestPath {
	r.Querystring.Del(key)
	return r
}

// Get querystring param.
func (r requestPath) Get(key string) string {
	return r.Querystring.Get(key)
}

// Set querystring param.
func (r *requestPath) Set(key, value string) *requestPath {
	r.Querystring.Set(key, value)
	return r
}

// Retrieve request path.
func (r requestPath) String() string {
	return (&url.URL{
		Path:     string(r.Endpoint),
		RawQuery: r.Querystring.Encode(),
	}).String()
}

// Get querystring param values.
func (r requestPath) Values(key string) []string {
	return r.Querystring[key]
}

// Make new request path.
func newRequestPath(ep apiEndpoint) *requestPath {
	return &requestPath{
		Endpoint:    ep,
		Querystring: make(url.Values),
	}
}
