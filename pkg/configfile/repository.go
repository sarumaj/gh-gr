package configfile

// Repository holds a repository URL and its local directory equivalent.
type Repository struct {
	URL       string `json:"URL" yaml:"URL"`
	Directory string `json:"directory" yaml:"directory"`
	Branch    string `json:"branch" yaml:"branch"`
	ParentURL string `json:"parentURL,omitempty" yaml:"parentURL,omitempty"`
	Public    bool   `json:"public,omitempty" yaml:"public,omitempty"`
	Size      string `json:"size" yaml:"size"`
}

type Repositories []Repository

// Append repository (only if not present).
func (r *Repositories) Append(repo Repository) {
	if !r.Has(repo) {
		*r = append(*r, repo)
	}
}

// Check if repo is enlisted (URL is considered to be unique).
func (r Repositories) Has(repo Repository) bool {
	for _, own := range r {
		if own.URL == repo.URL {
			return true
		}
	}

	return false
}

// Get the name of the repository with the longest name.
func (r Repositories) LongestName() string {
	var name string
	for _, own := range r {
		if len(own.Directory) > len(name) {
			name = own.Directory
		}
	}

	return name
}
