package configfile

import (
	"encoding/xml"
	"reflect"
)

// Repository holds a repository URL and its local directory equivalent.
type Repository struct {
	XMLName   xml.Name `xml:"repository" toml:"-" json:"-" yaml:"-"`
	URL       string   `xml:"url" toml:"url" json:"url" yaml:"url"`
	Directory string   `xml:"directory,attr" toml:"directory" json:"directory" yaml:"directory"`
	Branch    string   `xml:"branch,attr" toml:"branch" json:"branch" yaml:"branch"`
	ParentURL string   `xml:"parentUrl,omitempty" toml:"parentUrl,omitempty" json:"parentUrl,omitempty" yaml:"parentUrl,omitempty"`
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
