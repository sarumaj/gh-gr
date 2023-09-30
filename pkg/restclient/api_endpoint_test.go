package client

import "testing"

func TestApiEndpointFormat(t *testing.T) {
	type args struct {
		ep     apiEndpoint
		params map[string]any
	}
	for _, tt := range []struct {
		name string
		args args
		want apiEndpoint
	}{
		{"test#1", args{"parent/child", map[string]any{}}, "parent/child"},
		{"test#2", args{"parent/{daughter}", map[string]any{"daughter": "son"}}, "parent/son"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.ep.Format(tt.args.params)
			if got != tt.want {
				t.Errorf(`(apiEndpoint).Format(%v) failed: got: %q, want: %q`, tt.args.params, got, tt.want)
			}
		})
	}
}
