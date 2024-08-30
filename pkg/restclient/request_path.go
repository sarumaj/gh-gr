package restclient

import (
	"fmt"
	"net/url"
	"strings"
)

// Helper for building request paths.
type requestPath struct {
	Endpoint    apiEndpoint
	Querystring url.Values
	Validations map[string][]string
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

// Register a querystring param validation.
func (r *requestPath) Register(key string, values ...string) *requestPath {
	if r.Validations == nil {
		r.Validations = make(map[string][]string)
	}

	r.Validations[key] = values
	return r
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

// Validate querystring params.
func (r requestPath) Validate() error {
	if len(r.Validations) == 0 {
		return nil
	}

	for key, values := range r.Querystring {
		validValues, ok := r.Validations[key]
		if !ok {
			continue
		}

		valid := false
		for _, validValue := range validValues {
			validAll := true
			for _, value := range values {
				validAll = validAll && value == validValue && value != ""
			}
			valid = valid || validAll
		}

		if !valid {
			return fmt.Errorf("invalid filter value for %s, allowed values: %s", key, strings.Join(values, ", "))
		}
	}

	return nil
}

// Make new request path.
func newRequestPath(ep apiEndpoint) *requestPath {
	return newRequestPathWithValidations(ep, nil)
}

// Make new request path with validations.
func newRequestPathWithValidations(ep apiEndpoint, validations map[string][]string) *requestPath {
	return &requestPath{
		Endpoint:    ep,
		Querystring: make(url.Values),
		Validations: validations,
	}
}
