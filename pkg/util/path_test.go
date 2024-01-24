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

func TestStripPathPrefix(t *testing.T) {
	type args struct {
		path        string
		keepParents uint
	}

	for _, tt := range []struct {
		name string
		args args
		want string
	}{
		{"test#1", args{"", 0}, ""},
		{"test#2", args{"/dir1/dir2/item", 1}, "dir2/item"},
		{"test#3", args{"/dir1/dir2/item", 2}, "dir1/dir2/item"},
		{"test#4", args{"/dir1/../dir2/item", 1}, "dir2/item"},
		{"test#5", args{"/dir1/../dir2/item", 10}, "/dir2/item"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := StripPathPrefix(tt.args.path, tt.args.keepParents)
			if got != tt.want {
				t.Errorf(`StripPathPrefix(%q, %d) failed: got: %q, want: %q`, tt.args.path, tt.args.keepParents, got, tt.want)
			}
		})
	}
}
