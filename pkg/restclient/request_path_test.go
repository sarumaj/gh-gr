package restclient

import "testing"

func TestRequestPathStringify(t *testing.T) {
	for _, tt := range []struct {
		name string
		arg  *requestPath
		want string
	}{
		{"test#1", newRequestPath(""), ""},
		{"test#2", newRequestPath("parent/child"), "parent/child"},
		{"test#3", newRequestPath("parent/child").Add("param1", "val1", "val2"), "parent/child?param1=val1&param1=val2"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.arg.String()
			if got != tt.want {
				t.Errorf(`(requestPath).String() failed: got: %q, want %q`, got, tt.want)
			}
		})
	}
}
