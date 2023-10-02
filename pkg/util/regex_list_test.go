package util

import "testing"

func TestRegexListMatch(t *testing.T) {
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
			got := RegexList(tt.args.list).Match(tt.args.target)
			if got != tt.want {
				t.Errorf(`RegexList(%v).Match(%q) failed: got: %t, want: %t`, tt.args.list, tt.args.target, got, tt.want)
			}
		})
	}

}
