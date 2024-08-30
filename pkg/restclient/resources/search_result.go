package resources

type SearchResult[T any] struct {
	TotalCount        int  `json:"total_count"`
	IncompleteResults bool `json:"incomplete_results"`
	Items             []T  `json:"items"`
}
