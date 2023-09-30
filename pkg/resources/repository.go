package resources

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Repository struct {
	ID                  int                `json:"id"`
	NodeID              string             `json:"node_id"`
	Name                string             `json:"name"`
	FullName            string             `json:"full_name"`
	Owner               Owner              `json:"owner"`
	Private             bool               `json:"private"`
	HTMLURL             string             `json:"html_url"`
	Description         string             `json:"description"`
	Fork                bool               `json:"fork"`
	URL                 string             `json:"url"`
	ArchiveURL          string             `json:"archive_url"`
	AssigneesURL        string             `json:"assignees_url"`
	BlobsURL            string             `json:"blobs_url"`
	BranchesURL         string             `json:"branches_url"`
	CollaboratorsURL    string             `json:"collaborators_url"`
	CommentsURL         string             `json:"comments_url"`
	CommitsURL          string             `json:"commits_url"`
	CompareURL          string             `json:"compare_url"`
	ContentsURL         string             `json:"contents_url"`
	ContributorsURL     string             `json:"contributors_url"`
	DeploymentsURL      string             `json:"deployments_url"`
	DownloadsURL        string             `json:"downloads_url"`
	EventsURL           string             `json:"events_url"`
	ForksURL            string             `json:"forks_url"`
	GitCommitsURL       string             `json:"git_commits_url"`
	GitRefsURL          string             `json:"git_refs_url"`
	GitTagsURL          string             `json:"git_tags_url"`
	GitURL              string             `json:"git_url"`
	IssueCommentURL     string             `json:"issue_comment_url"`
	IssueEventsURL      string             `json:"issue_events_url"`
	IssuesURL           string             `json:"issues_url"`
	KeysURL             string             `json:"keys_url"`
	LabelsURL           string             `json:"labels_url"`
	LanguagesURL        string             `json:"languages_url"`
	MergesURL           string             `json:"merges_url"`
	MilestonesURL       string             `json:"milestones_url"`
	NotificationsURL    string             `json:"notifications_url"`
	PullsURL            string             `json:"pulls_url"`
	ReleasesURL         string             `json:"releases_url"`
	SSHURL              string             `json:"ssh_url"`
	StargazersURL       string             `json:"stargazers_url"`
	StatusesURL         string             `json:"statuses_url"`
	SubscribersURL      string             `json:"subscribers_url"`
	SubscriptionURL     string             `json:"subscription_url"`
	TagsURL             string             `json:"tags_url"`
	TeamsURL            string             `json:"teams_url"`
	TreesURL            string             `json:"trees_url"`
	CloneURL            string             `json:"clone_url"`
	MirrorURL           string             `json:"mirror_url"`
	HooksURL            string             `json:"hooks_url"`
	SvnURL              string             `json:"svn_url"`
	Homepage            string             `json:"homepage"`
	Language            any                `json:"language"`
	ForksCount          int                `json:"forks_count"`
	Forks               int                `json:"forks"`
	StargazersCount     int                `json:"stargazers_count"`
	WatchersCount       int                `json:"watchers_count"`
	Watchers            int                `json:"watchers"`
	Size                int                `json:"size"`
	DefaultBranch       string             `json:"default_branch"`
	OpenIssuesCount     int                `json:"open_issues_count"`
	OpenIssues          int                `json:"open_issues"`
	IsTemplate          bool               `json:"is_template"`
	Topics              []string           `json:"topics"`
	HasIssues           bool               `json:"has_issues"`
	HasProjects         bool               `json:"has_projects"`
	HasWiki             bool               `json:"has_wiki"`
	HasPages            bool               `json:"has_pages"`
	HasDownloads        bool               `json:"has_downloads"`
	HasDiscussions      bool               `json:"has_discussions"`
	Archived            bool               `json:"archived"`
	Disabled            bool               `json:"disabled"`
	Visibility          string             `json:"visibility"`
	PushedAt            time.Time          `json:"pushed_at"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
	Permissions         Permissions        `json:"permissions"`
	AllowRebaseMerge    bool               `json:"allow_rebase_merge"`
	TemplateRepository  TemplateRepository `json:"template_repository"`
	TempCloneToken      string             `json:"temp_clone_token"`
	AllowSquashMerge    bool               `json:"allow_squash_merge"`
	AllowAutoMerge      bool               `json:"allow_auto_merge"`
	DeleteBranchOnMerge bool               `json:"delete_branch_on_merge"`
	AllowMergeCommit    bool               `json:"allow_merge_commit"`
	SubscribersCount    int                `json:"subscribers_count"`
	NetworkCount        int                `json:"network_count"`
	License             License            `json:"license"`
	Organization        Organization       `json:"organization"`
	Parent              Parent             `json:"parent,omitempty"` // Present if repository is a fork.
	Source              Source             `json:"source,omitempty"` // Present if repository is a fork.
}

type Repositories []Repository

func (r Repositories) Print(out io.Writer) {
	if out == nil {
		out = os.Stdout
	}

	for _, repo := range r {
		fmt.Fprintf(out, " - %s\n", repo.FullName)
	}

	fmt.Fprintf(out, "Total number of repositories: %d\n", len(r))
}
