package util

import "testing"

func TestGetHostnameFromPath(t *testing.T) {
	for _, tt := range []struct {
		name string
		args string
		want string
	}{
		{"test#1", "https://example.com/endpoint?q=search", "example.com"},
		{"test#2", "https://example.com/endpoint", "example.com"},
		{"test#3", "https://example.com:443", "example.com"},
		{"test#4", "example.com/endpoint", "example.com"},
		{"test#5", "//example.com/endpoint", "example.com"},
		{"test#6", "http://example.com:443", "example.com"},
		{"test#7", "http://example.com:443", "example.com"},
		{"test#8", "http://example.com:443/endpoint", "example.com"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := GetHostnameFromPath(tt.args)
			if got != tt.want {
				t.Errorf(`GetHostnameFromPath(%q) failed: got: %q, want: %q`, tt.args, got, tt.want)
			}
		})
	}
}
