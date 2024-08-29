package resources

// Branch represents a GitHub branch.
type Branch struct {
	Label string     `json:"label"`
	Ref   string     `json:"ref"`
	Sha   string     `json:"sha"`
	User  User       `json:"user"`
	Repo  Repository `json:"repo"`
}
