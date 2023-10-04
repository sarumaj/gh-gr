package configfile

import "reflect"

// Repository holds a repository URL and its local directory equivalent.
type Repository struct {
	URL       string `yaml:"url"`
	Directory string `yaml:"directory"`
	Branch    string `yaml:"branch"`
	ParentURL string `yaml:"parentUrl,omitempty"`
}

type Repositories []Repository

func (r Repositories) AppendRepository(repo Repository) Repositories {
	if !r.HasRepository(repo) {
		r = append(r, repo)
	}

	return r
}

func (r Repositories) HasRepository(repo Repository) bool {
	for _, own := range r {
		if reflect.DeepEqual(own, repo) {
			return true
		}
	}

	return false
}
