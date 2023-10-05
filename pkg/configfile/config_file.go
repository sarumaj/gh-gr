package configfile

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	auth "github.com/cli/go-gh/v2/pkg/auth"
	config "github.com/cli/go-gh/v2/pkg/config"
	prompter "github.com/cli/go-gh/v2/pkg/prompter"
	term "github.com/cli/go-gh/v2/pkg/term"
	color "github.com/fatih/color"
	util "github.com/sarumaj/gh-gr/pkg/util"
	yaml "gopkg.in/yaml.v2"
)

const configKey = "gr.conf"

const (
	ConfigNotFound       = "No configuration found. Make sure to run 'init' to create initial configuration."
	ConfigShouldNotExist = "Configuration already exists. " +
		"Please run 'update' if you want to update your settings. " +
		"Alternatively, run 'remove' if you want to setup from scratch once again."
)

var urlRegex = regexp.MustCompile(`(?P<Schema>[^:]+://)(?P<Creds>[^@]+@)?(?P<Hostpath>.+)`)

// Configuration holds gr configuration data
type Configuration struct {
	Username       string        `yaml:"username"`
	Fullname       string        `yaml:"fullname"`
	Email          string        `yaml:"email,omitempty"`
	BaseDirectory  string        `yaml:"baseDirectory"`
	BaseURL        string        `yaml:"baseURL"`
	Concurrency    uint          `yaml:"concurrency"`
	SubDirectories bool          `yaml:"subDirectories"`
	Verbose        bool          `yaml:"verbose"`
	Timeout        time.Duration `yaml:"timeout"`
	Excluded       []string      `yaml:"exluded,omitempty"`
	Included       []string      `yaml:"included,omitempty"`
	Repositories   Repositories  `yaml:"repositories"`
}

func ConfigurationExists() bool {
	c, err := config.Read()
	if err != nil {
		return false
	}

	raw, err := c.Get([]string{configKey})
	return err == nil && len(raw) > 0
}

func Load() *Configuration {
	var conf Configuration

	c, err := config.Read()
	util.FatalIfError(err)

	content, err := c.Get([]string{configKey})
	util.FatalIfError(err)

	bar := binaryProgressbar().Describe(util.CheckColors(color.BlueString, "Loading..."))
	util.FatalIfError(yaml.NewDecoder(io.TeeReader(strings.NewReader(content), bar)).Decode(&conf))
	_ = bar.Clear()

	return &conf
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

func (conf Configuration) Copy() *Configuration {
	n := &Configuration{
		Username:       conf.Username,
		Fullname:       conf.Fullname,
		Email:          conf.Email,
		BaseDirectory:  conf.BaseDirectory,
		BaseURL:        conf.BaseURL,
		Concurrency:    conf.Concurrency,
		SubDirectories: conf.SubDirectories,
		Verbose:        conf.Verbose,
		Timeout:        conf.Timeout,
	}

	if conf.Excluded != nil {
		n.Excluded = make([]string, 0)
		_ = copy(n.Excluded, conf.Excluded)
	}

	if conf.Included != nil {
		n.Included = make([]string, 0)
		_ = copy(n.Included, conf.Included)
	}

	if conf.Repositories != nil {
		n.Repositories = make(Repositories, 0)
		_ = copy(n.Repositories, conf.Repositories)
	}

	return n
}

func (conf Configuration) Display() {
	util.FatalIfError(yaml.NewEncoder(os.Stdout).Encode(conf))
}

func (conf Configuration) GetToken() string {
	host := conf.BaseURL

	if parsed, err := url.Parse(conf.BaseURL); err == nil {
		host = parsed.Hostname()
	}

	token, _ := auth.TokenForHost(host)
	return token
}

func (conf Configuration) Remove(purge bool) {
	c, err := config.Read()
	util.FatalIfError(err)

	util.FatalIfError(c.Remove([]string{configKey}))
	util.FatalIfError(config.Write(c))

	fmt.Println(util.CheckColors(color.GreenString, "Configuration removed."))

	if !purge {
		return
	}

	if term.IsTerminal(os.Stdout) && term.IsTerminal(os.Stderr) {
		confirm, err := prompter.New(os.Stdin, os.Stdout, os.Stderr).
			Confirm(
				util.CheckColors(
					color.RedString,
					"DANGER!!! ",
				)+"You will delete all local repositories! Are you sure?",
				false,
			)
		util.FatalIfError(err)

		if !confirm {
			return
		}
	}

	bar := util.NewProgressbar(len(conf.Repositories))
	for _, repo := range conf.Repositories {
		bar.Describe(util.CheckColors(color.RedString, "Removing %s...", repo.Directory))
		util.FatalIfError(os.RemoveAll(repo.Directory))
		bar.Inc()
	}

	if conf.BaseDirectory != "." {
		util.FatalIfError(os.RemoveAll(conf.BaseDirectory))
	}

	fmt.Println(util.CheckColors(color.GreenString, "Successfully removed repositories from local filesystem."))
}

func (conf Configuration) Save() {
	c, err := config.Read()
	util.FatalIfError(err)

	buffer := bytes.NewBuffer(nil)
	bar := binaryProgressbar().Describe(util.CheckColors(color.BlueString, "Saving..."))
	util.FatalIfError(yaml.NewEncoder(io.MultiWriter(buffer, bar)).Encode(conf))
	_ = bar.Clear()

	c.Set([]string{configKey}, buffer.String())
	util.FatalIfError(config.Write(c))

	fmt.Println(util.CheckColors(color.GreenString, "Configuration saved. You can now pull your repositories."))
}

func binaryProgressbar() *util.Progressbar {
	return util.NewProgressbar(
		-1,
		util.EnableColorCodes(util.UseColors()),
		util.SetWidth(10),
		util.ShowBytes(true),
		util.SetRenderBlankState(true),
		util.ClearOnFinish(),
		util.ShowCount(),
	)
}
