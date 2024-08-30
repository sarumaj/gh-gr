package restclient

import (
	"net/http"
	"testing"

	"github.com/sarumaj/gh-gr/v2/pkg/restclient/resources"
)

func TestConsolidate(t *testing.T) {
	{
		got := consolidate[string]([]string{"a", "b"}, []string{"c", "d"})
		if len(got) != 4 {
			t.Errorf("consolidate[string, []string]() failed: got: %v, want: %v", got, []string{"a", "b", "c", "d"})
		}
	}
	{
		got := consolidate[string](
			resources.SearchResult[string]{Items: []string{"a", "b"}},
			resources.SearchResult[string]{Items: []string{"c", "d"}},
		)
		if len(got.Items) != 4 {
			t.Errorf("consolidate[string, resources.SearchResult[string]]() failed: got: %v, want: %v", got.Items, []string{"a", "b", "c", "d"})
		}
	}
}

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
