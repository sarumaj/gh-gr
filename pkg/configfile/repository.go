package configfile

// Repository holds a repository URL and its local directory equivalent.
type Repository struct {
	URL       string `yaml:"url"`
	Directory string `yaml:"directory"`
	Branch    string `yaml:"branch"`
	ParentURL string `yaml:"parentUrl"`
}
