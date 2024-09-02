package configfile

import (
	"fmt"
	"time"

	"github.com/sarumaj/gh-gr/v2/pkg/restclient/resources"
)

// PullRequest represents a GitHub pull request.
type PullRequest struct {
	Assignees  []string         `json:"assignee" yaml:"assignee"`
	Author     string           `json:"author" yaml:"author"`
	Base       string           `json:"base" yaml:"base"`
	ClosedAt   time.Time        `json:"closed_at" yaml:"closed_at"`
	Error      PullRequestError `json:"error" yaml:"error"`
	Head       string           `json:"head" yaml:"head"`
	Labels     []string         `json:"labels" yaml:"labels"`
	Number     int              `json:"number" yaml:"number"`
	Repository string           `json:"repository" yaml:"repository"`
	State      string           `json:"state" yaml:"state"`
	Title      string           `json:"title" yaml:"title"`
	URL        string           `json:"URL" yaml:"URL"`
}

// PullRequestError represents an error that occurred while processing a PullRequest.
type PullRequestError string

// Error returns the error message.
func (err PullRequestError) Error() string {
	return fmt.Sprintf("failed: %s", string(err))
}

// PullRequestList is a list of PullRequest.
type PullRequestList []PullRequest

// Append appends a PullRequest to the PullRequestList if it is not already in the list.
func (prl *PullRequestList) Append(pr PullRequest) {
	if !prl.Has(pr) {
		*prl = append(*prl, pr)
	}
}

// Browse allows to open the URLs of the PullRequests in the default browser.
func (prl PullRequestList) Browse() {
	links := make([]string, len(prl))
	for i, pr := range prl {
		links[i] = pr.URL
	}

	OpenLins(links)
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

// Status returns the status of the PullRequest.
func (prl PullRequest) Status() any {
	if prl.Error != "" {
		return prl.Error
	}

	return prl.State
}

// PullRequestFromResponse creates a PullRequest from a GitHub API response.
func PullRequestFromResponse(response resources.PullRequest) PullRequest {
	pr := PullRequest{
		Assignees:  []string{response.Assignee.Login},
		Author:     response.User.Login,
		Base:       response.Base.Ref,
		ClosedAt:   response.ClosedAt,
		Head:       response.Head.Ref,
		Number:     response.Number,
		Repository: response.Repository,
		State:      response.State,
		Title:      response.Title,
		URL:        response.HTMLURL,
	}

	if len(pr.Title) > 40 {
		pr.Title = pr.Title[:37] + "..."
	}

	appendUnique := func(slice []string, in string) []string {
		for _, item := range slice {
			if item == in {
				return slice
			}
		}
		return append(slice, in)
	}

	for _, assignee := range response.Assignees {
		pr.Assignees = appendUnique(pr.Assignees, assignee.Login)
	}

	for _, label := range response.Labels {
		pr.Labels = appendUnique(pr.Labels, label.Name)
	}

	return pr
}
