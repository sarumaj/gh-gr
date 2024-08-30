package restclient

import (
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Constants for rate limiting.
const (
	primaryRateLimit   = 5000      // Maximum number of requests per hour
	secondaryRateLimit = 100       // Maximum of 100 concurrent points
	rateLimitWindow    = time.Hour // Time window for the primary rate limit
)

// Default transport with rate limiting applied.
var defaultTransport = newThrottledTransport()

// throttledTransport wraps an http.RoundTripper to throttle requests.
type throttledTransport struct {
	Transport    http.RoundTripper
	requestTimes []time.Time
	points       int64
	mu           sync.Mutex
	rateReset    time.Time
	retry        bool
}

// calculatePointCost determines the cost of a request based on its HTTP method.
func (t *throttledTransport) calculatePointCost(method string) int64 {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return 1
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return 5
	default:
		return 1
	}
}

// handlePrimaryRateLimit enforces the primary rate limit window.
func (t *throttledTransport) handlePrimaryRateLimit() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rateLimitWindow)
	validRequests := 0

	for _, ts := range t.requestTimes {
		if ts.After(cutoff) {
			break
		}
		validRequests++
	}

	t.requestTimes = t.requestTimes[validRequests:]

	if len(t.requestTimes) >= primaryRateLimit {
		sleepDuration := t.requestTimes[0].Add(rateLimitWindow).Sub(now)
		time.Sleep(sleepDuration)
	}

	t.requestTimes = append(t.requestTimes, time.Now())
}

// shouldRetry determines if a request should be retried based on rate limit headers.
func (t *throttledTransport) shouldRetry(req *http.Request) (time.Duration, error) {
	remainingStr := req.Header.Get("X-RateLimit-Remaining")
	if remainingStr == "" {
		return 0, nil
	}

	remaining, err := strconv.Atoi(remainingStr)
	if err != nil {
		return 0, err
	}

	if remaining > 0 {
		return 0, nil
	}

	resetStr := req.Header.Get("X-RateLimit-Reset")
	if resetStr == "" {
		return 0, nil
	}

	reset, err := strconv.Atoi(resetStr)
	if err != nil {
		return 0, err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.rateReset = time.Unix(int64(reset), 0).Add(time.Second)
	return time.Until(t.rateReset), nil
}

// waitForSecondaryRateLimit waits until the secondary rate limit allows new requests.
func (t *throttledTransport) waitForSecondaryRateLimit(pointCost int64) {
	for {
		t.mu.Lock()
		if atomic.LoadInt64(&t.points)+pointCost < secondaryRateLimit {
			t.mu.Unlock()
			break
		}
		t.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
}

// RoundTrip implements the http.RoundTripper interface, managing request throttling.
func (t *throttledTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.retry {
		waitDuration, err := t.shouldRetry(req)
		if err != nil {
			return nil, err
		}

		if waitDuration > 0 {
			time.Sleep(waitDuration)
		}
	}

	pointCost := t.calculatePointCost(req.Method)
	t.waitForSecondaryRateLimit(pointCost)
	atomic.AddInt64(&t.points, pointCost)
	t.handlePrimaryRateLimit()
	resp, err := t.Transport.RoundTrip(req)
	atomic.AddInt64(&t.points, -pointCost)

	return resp, err
}

// SetRetry sets the retry flag to enable or disable retrying.
func (t *throttledTransport) SetRetry(retry bool) {
	t.mu.Lock()
	t.retry = retry
	t.mu.Unlock()
}

// SetTransport sets the underlying transport.
func (t *throttledTransport) SetTransport(rt http.RoundTripper) {
	if rt == nil {
		return
	}

	t.mu.Lock()
	t.Transport = rt
	t.mu.Unlock()
}

// newThrottledTransport creates a new ThrottledTransport with a default transport.
func newThrottledTransport() *throttledTransport {
	return &throttledTransport{
		Transport: http.DefaultTransport,
		points:    0,
	}
}
