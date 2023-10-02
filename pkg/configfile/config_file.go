package configfile

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	auth "github.com/cli/go-gh/v2/pkg/auth"
	config "github.com/cli/go-gh/v2/pkg/config"
	prompter "github.com/cli/go-gh/v2/pkg/prompter"
	util "github.com/sarumaj/gh-gr/pkg/util"
	yaml "gopkg.in/yaml.v2"
)

const configKey = "gr.conf"

var urlRegex = regexp.MustCompile(`(?P<Schema>[^:]+://)(?P<Creds>[^@]+@)?(?P<Hostpath>.+)`)

// Configuration holds gr configuration data
type Configuration struct {
	Username       string       `yaml:"username"`
	Fullname       string       `yaml:"fullname"`
	Email          string       `yaml:"Email"`
	BaseDirectory  string       `yaml:"baseDirectory"`
	BaseURL        string       `yaml:"baseURL"`
	Concurrency    uint         `yaml:"concurrency"`
	SubDirectories bool         `yaml:"subDirectories"`
	Excluded       []string     `yaml:"exluded"`
	Included       []string     `yaml:"included"`
	Repositories   []Repository `yaml:"repositories"`
	Verbose        bool         `yaml:"-"`
}

func (conf Configuration) Authenticate(targetURL *string) {
	if targetURL == nil || *targetURL == "" || !urlRegex.MatchString(*targetURL) {
		return
	}

	parsed, err := url.Parse(urlRegex.ReplaceAllString(
		*targetURL,
		fmt.Sprintf("${Schema}%s:%s@${Hostpath}", conf.Username, conf.GetToken()),
	))
	if err != nil {
		return
	}

	*targetURL = parsed.String()
}

func (Configuration) Exists() bool {
	return util.PathExists(configKey)
}

func (conf Configuration) GetToken() string {
	token, _ := auth.TokenForHost(conf.BaseURL)
	return token
}

func Load() *Configuration {
	c, err := config.Read()
	util.FatalIfError(err)

	content, err := c.Get([]string{configKey})
	util.FatalIfError(err)

	var conf Configuration
	util.FatalIfError(yaml.Unmarshal([]byte(content), &conf))

	return &conf
}

func (conf Configuration) Remove(purge bool) {
	c, err := config.Read()
	util.FatalIfError(err)

	util.FatalIfError(c.Remove([]string{configKey}))

	fmt.Println("Configuration removed.")

	if !purge {
		return
	}

	confirm, err := prompter.New(os.Stdin, os.Stdout, os.Stderr).
		Confirm("DANGER!!! You will delete all local repositories! Are you sure?", false)
	util.FatalIfError(err)

	if !confirm {
		return
	}

	for _, repo := range conf.Repositories {
		util.FatalIfError(os.RemoveAll(filepath.Join(conf.BaseDirectory, repo.Directory)))
	}

	if conf.BaseDirectory != "." {
		util.FatalIfError(os.RemoveAll(conf.BaseDirectory))
	}
}

func (conf Configuration) Save() {
	c, err := config.Read()
	util.FatalIfError(err)

	content, err := yaml.Marshal(conf)
	util.FatalIfError(err)

	c.Set([]string{configKey}, string(content))
	util.FatalIfError(config.Write(c))

	fmt.Println("Configuration saved. You can now pull your repositories.")
}
