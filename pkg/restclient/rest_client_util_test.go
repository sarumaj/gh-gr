package restclient

import (
	"net/http"
	"testing"
)

func TestGetLastPage(t *testing.T) {
	tests := []struct {
		name string
		args http.Header
		want int
	}{
		{"test#1",
			http.Header{"Link": {`<https://api.github.com/anything?per_page=2&page=2>; rel="next", ` +
				`<https://api.github.com/anything?per_page=2&page=7715>; rel="last"`}},
			7715,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotLimit := getLastPage(tt.args); gotLimit != tt.want {
				t.Errorf("getLastPage() failed: got: %d, want: %d", gotLimit, tt.want)
			}
		})
	}
}
