package resources

// PullRequestLinks represents the links associated with a GitHub pull request.
type PullRequestLinks struct {
	Self           Href `json:"self"`
	HTML           Href `json:"html"`
	Issue          Href `json:"issue"`
	Comments       Href `json:"comments"`
	ReviewComments Href `json:"review_comments"`
	ReviewComment  Href `json:"review_comment"`
	Commits        Href `json:"commits"`
	Statuses       Href `json:"statuses"`
}
