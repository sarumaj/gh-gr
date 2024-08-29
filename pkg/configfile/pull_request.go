package configfile

// PullRequest represents a GitHub pull request.
type PullRequest struct {
	Title      string   `json:"title" yaml:"title"`
	URL        string   `json:"URL" yaml:"URL"`
	Number     int      `json:"number" yaml:"number"`
	State      string   `json:"state" yaml:"state"`
	Author     string   `json:"author" yaml:"author"`
	Assignees  []string `json:"assignee" yaml:"assignee"`
	Repository string   `json:"repository" yaml:"repository"`
	Labels     []string `json:"labels" yaml:"labels"`
}

// PullRequestList is a list of PullRequest.
type PullRequestList []PullRequest

// Append appends a PullRequest to the PullRequestList if it is not already in the list.
func (prl *PullRequestList) Append(pr PullRequest) {
	if !prl.Has(pr) {
		*prl = append(*prl, pr)
	}
}

// Has returns true if the PullRequestList contains the given PullRequest.
func (prl PullRequestList) Has(pr PullRequest) bool {
	for _, own := range prl {
		if own.URL == pr.URL {
			return true
		}
	}

	return false
}

// LongestTitle returns the title of the PullRequest with the longest title.
func (prl PullRequestList) LongestTitle() string {
	var name string
	for _, own := range prl {
		if len(own.Title) > len(name) {
			name = own.Title
		}
	}

	return name
}
