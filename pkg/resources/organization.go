package resources

import (
	"fmt"
	"io"
	"os"
)

type Organization struct {
	Login            string `json:"login"`
	ID               int    `json:"id"`
	NodeID           string `json:"node_id"`
	URL              string `json:"url"`
	ReposURL         string `json:"repos_url"`
	EventsURL        string `json:"events_url"`
	HooksURL         string `json:"hooks_url"`
	IssuesURL        string `json:"issues_url"`
	MembersURL       string `json:"members_url"`
	PublicMembersURL string `json:"public_members_url"`
	AvatarURL        string `json:"avatar_url"`
	Description      string `json:"description"`
}

type Organizations []Organization

func (o Organizations) Print(out io.Writer) {
	if out == nil {
		out = os.Stdout
	}

	for _, org := range o {
		fmt.Fprintf(out, " - %s\n", org.Login)
	}

	fmt.Fprintf(out, "Total number of organizations: %d\n", len(o))
}
