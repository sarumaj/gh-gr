package resources

type RateResources struct {
	Core                      Rate `json:"core"`
	Search                    Rate `json:"search"`
	Graphql                   Rate `json:"graphql"`
	IntegrationManifest       Rate `json:"integration_manifest"`
	SourceImport              Rate `json:"source_import"`
	CodeScanningUpload        Rate `json:"code_scanning_upload"`
	ActionsRunnerRegistration Rate `json:"actions_runner_registration"`
	Scim                      Rate `json:"scim"`
	DependencySnapshots       Rate `json:"dependency_snapshots"`
	CodeSearch                Rate `json:"code_search"`
}
