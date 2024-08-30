package restclient

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const primaryRateLimit = 5000     // Maximum number of requests per hour
const secondaryRateLimit = 100    // Maximum of 100 concurrent points
const rateLimitWindow = time.Hour // Time window for the primary rate limit

var defaultTransport = newThrottledTransport()

// Wrap an http.RoundTripper to throttle requests.
type throttledTransport struct {
	Transport    http.RoundTripper
	requestTimes []time.Time
	points       int64
	mu           sync.Mutex
}

// Implement the RoundTrip method of the http.RoundTripper interface.
func (t *throttledTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var pointCost int64
	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		pointCost = 1
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		pointCost = 5
	}

	for {
		t.mu.Lock()
		currentPoints := t.points
		t.mu.Unlock()

		if currentPoints+pointCost < secondaryRateLimit {
			break
		}
		time.Sleep(10 * time.Millisecond) // Sleep to avoid busy-waiting
	}

	// increment the points counter
	_ = atomic.AddInt64(&t.points, pointCost)

	t.mu.Lock()

	// count requests within the rate limit window
	now := time.Now()
	cutoff := now.Add(-rateLimitWindow)
	validRequests := 0
	for _, ts := range t.requestTimes {
		if ts.After(cutoff) {
			break
		}
		validRequests++
	}
	// remove requests outside the rate limit window
	t.requestTimes = t.requestTimes[validRequests:]

	// wait if the rate limit is exceeded
	if len(t.requestTimes) >= primaryRateLimit {
		sleepDuration := t.requestTimes[0].Add(rateLimitWindow).Sub(now)
		time.Sleep(sleepDuration)
	}

	// add the current request to the list
	t.requestTimes = append(t.requestTimes, time.Now())
	t.mu.Unlock()

	// perform the request
	resp, err := t.Transport.RoundTrip(req)

	// decrement the points counter
	_ = atomic.AddInt64(&t.points, -pointCost)
	return resp, err
}

// Create a new ThrottledTransport with a specified rate.
func newThrottledTransport() *throttledTransport {
	return &throttledTransport{
		Transport: http.DefaultTransport, // You can use custom transports as well
		points:    0,
	}
}
