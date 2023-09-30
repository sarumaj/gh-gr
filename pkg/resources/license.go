package resources

type License struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	SpdxID  string `json:"spdx_id"`
	NodeID  string `json:"node_id"`
	HTMLURL string `json:"html_url,omitempty"`
}
