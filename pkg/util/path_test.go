package util

import (
	"runtime"
	"testing"
)

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
		{"test#9", "https://username:token@example.com/organization/repository.git", "example.com"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := GetHostnameFromPath(tt.args)
			if got != tt.want {
				t.Errorf(`GetHostnameFromPath(%q) failed: got: %q, want: %q`, tt.args, got, tt.want)
			}
		})
	}
}

func TestPathSanitize(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skipf("Test should run only on windows")
	}

	for _, tt := range []struct {
		name string
		args string
		want string
	}{
		{"test#1", "C:\\Users\\admin\\github", "/Users/admin/github"},
		{"test#2", "\\home\\dir\\github\\", "/home/dir/github"},
		{"test#3", "home\\dir\\github\\", "home/dir/github"},
		{"test#4", "D:\\Users\\admin\\github", "D:/Users/admin/github"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args
			PathSanitize(&got)
			if got != tt.want {
				t.Errorf(`PathSanitize(&%q) failed: got: %q, want: %q`, tt.args, got, tt.want)
			}
		})
	}
}
