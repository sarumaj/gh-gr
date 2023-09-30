package client

import (
	"net/url"
)

type requestPath struct {
	Endpoint    apiEndpoint
	Querystring url.Values
}

func (r *requestPath) Add(key string, values ...string) *requestPath {
	for _, value := range values {
		r.Querystring.Add(key, value)
	}

	return r
}

func (r *requestPath) Del(key string) *requestPath {
	r.Querystring.Del(key)
	return r
}

func (r requestPath) Get(key string) string {
	return r.Querystring.Get(key)
}

func (r *requestPath) Set(key, value string) *requestPath {
	r.Querystring.Set(key, value)
	return r
}

func (r requestPath) String() string {
	return (&url.URL{
		Path:     string(r.Endpoint),
		RawQuery: r.Querystring.Encode(),
	}).String()
}

func (r requestPath) Values(key string) []string {
	return r.Querystring[key]
}

func newRequestPath(ep apiEndpoint) *requestPath {
	return &requestPath{
		Endpoint:    ep,
		Querystring: make(url.Values),
	}
}
