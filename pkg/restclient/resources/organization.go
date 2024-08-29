package resources

import (
	"fmt"
	"time"
)

type Organization struct {
	Login                   string    `json:"login"`
	ID                      int       `json:"id"`
	NodeID                  string    `json:"node_id"`
	URL                     string    `json:"url"`
	ReposURL                string    `json:"repos_url"`
	EventsURL               string    `json:"events_url"`
	HooksURL                string    `json:"hooks_url"`
	IssuesURL               string    `json:"issues_url"`
	MembersURL              string    `json:"members_url"`
	PublicMembersURL        string    `json:"public_members_url"`
	AvatarURL               string    `json:"avatar_url"`
	Description             string    `json:"description"`
	Name                    string    `json:"name"`
	Company                 string    `json:"company"`
	Blog                    string    `json:"blog"`
	Location                string    `json:"location"`
	Email                   string    `json:"email"`
	IsVerified              bool      `json:"is_verified"`
	HasOrganizationProjects bool      `json:"has_organization_projects"`
	HasRepositoryProjects   bool      `json:"has_repository_projects"`
	PublicRepos             int       `json:"public_repos"`
	PublicGists             int       `json:"public_gists"`
	Followers               int       `json:"followers"`
	Following               int       `json:"following"`
	HTMLURL                 string    `json:"html_url"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	Type                    string    `json:"type"`
}

func (o Organization) String() string {
	return fmt.Sprintf(" - ID:%d\tLogin:%s\tURL:%s", o.ID, o.Login, o.URL)
}
