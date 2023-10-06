package configfile

import (
	"reflect"
)

// Repository holds a repository URL and its local directory equivalent.
type Repository struct {
	URL       string `json:"URL" yaml:"URL"`
	Directory string `json:"directory" yaml:"directory"`
	Branch    string `json:"branch" yaml:"branch"`
	ParentURL string `json:"parentURL,omitempty" yaml:"parentURL,omitempty"`
}

type Repositories []Repository

func (r *Repositories) Append(repo Repository) {
	if !r.Has(repo) {
		*r = append(*r, repo)
	}
}

func (r Repositories) Has(repo Repository) bool {
	for _, own := range r {
		if reflect.DeepEqual(own, repo) {
			return true
		}
	}

	return false
}

func (r Repositories) LongestName() string {
	var name string
	for _, own := range r {
		if len(own.Directory) > len(name) {
			name = own.Directory
		}
	}

	return name
}
