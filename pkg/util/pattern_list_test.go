package util

import (
	"testing"
	"time"
)

func TestPatternListGlobMatch(t *testing.T) {
	type args struct {
		list   []string
		target string
	}
	for _, tt := range []struct {
		name string
		args args
		want bool
	}{
		{"test#1", args{[]string{""}, ""}, true},
		{"test#2", args{[]string{`owner/*`}, "owner/subject"}, true},
		{"test#3", args{[]string{`owner/[a-zA-Z]*`}, "owner/!@$"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := PatternList(tt.args.list).GlobMatch(tt.args.target)
			if got != tt.want {
				t.Errorf(`GlobList(%v).Match(%q) failed: got: %t, want: %t`, tt.args.list, tt.args.target, got, tt.want)
			}
		})
	}

}

func TestPatternListRegexMatch(t *testing.T) {
	type args struct {
		list   []string
		target string
	}
	for _, tt := range []struct {
		name string
		args args
		want bool
	}{
		{"test#1", args{[]string{""}, ""}, true},
		{"test#2", args{[]string{`^owner/\w*$`}, "owner/subject"}, true},
		{"test#3", args{[]string{`^owner/\w+$`}, "owner/!@$"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := PatternList(tt.args.list).RegexMatch(tt.args.target, time.Second)
			if got != tt.want {
				t.Errorf(`RegexList(%v).Match(%q) failed: got: %t, want: %t`, tt.args.list, tt.args.target, got, tt.want)
			}
		})
	}
}
