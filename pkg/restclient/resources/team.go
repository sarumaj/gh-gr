package resources

import "time"

type Team struct {
	ID                  int          `json:"id"`
	NodeID              string       `json:"node_id"`
	URL                 string       `json:"url"`
	HTMLURL             string       `json:"html_url"`
	Name                string       `json:"name"`
	Slug                string       `json:"slug"`
	Description         string       `json:"description"`
	Privacy             string       `json:"privacy"`
	NotificationSetting string       `json:"notification_setting"`
	Permission          string       `json:"permission"`
	MembersURL          string       `json:"members_url"`
	RepositoriesURL     string       `json:"repositories_url"`
	Parent              any          `json:"parent"`
	MembersCount        int          `json:"members_count"`
	ReposCount          int          `json:"repos_count"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
	Organization        Organization `json:"organization"`
}
