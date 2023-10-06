package configfile

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	auth "github.com/cli/go-gh/v2/pkg/auth"
	config "github.com/cli/go-gh/v2/pkg/config"
	prompter "github.com/cli/go-gh/v2/pkg/prompter"
	term "github.com/cli/go-gh/v2/pkg/term"
	color "github.com/fatih/color"
	util "github.com/sarumaj/gh-gr/pkg/util"
	"github.com/sirupsen/logrus"
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
var loggerEntry = util.Logger.WithFields(logrus.Fields{"mod": "configfile"})

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
	c, err := config.Read()
	util.FatalIfError(err)

	content, err := c.Get([]string{configKey})
	util.FatalIfError(err)

	var conf Configuration
	bar := newBinaryProgressbar().Describe(util.CheckColors(color.BlueString, "Loading..."))
	util.FatalIfError(yaml.NewDecoder(io.TeeReader(strings.NewReader(content), bar)).Decode(&conf))
	_ = bar.Clear()

	return &conf
}

func (conf Configuration) Authenticate(targetURL *string) {
	logger := loggerEntry
	logger.Debugf("Authenticating URL: %v", targetURL)
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
	reader, writer := io.Pipe()

	go func() {
		defer writer.Close()
		util.FatalIfError(yaml.NewEncoder(writer).Encode(conf))
	}()

	interactive := term.IsTerminal(os.Stdout) &&
		term.IsTerminal(os.Stdin) &&
		util.UseColors()

	for iter, noLines, scanner := 0, 10, bufio.NewScanner(reader); scanner.Scan(); iter++ {
		fmt.Fprintln(os.Stdout, scanner.Text())

		if interactive && iter > 0 && iter%noLines == 0 {
			fmt.Fprint(os.Stdout, color.BlueString("(more):"))

			var in string
			fmt.Fscanln(os.Stdin, &in)
			fmt.Fprint(os.Stdout, "\033[1A\r") // move one line up and use carriage return to move to the beginning of line
			if strings.HasPrefix(strings.ToLower(in), "q") {
				fmt.Fprintln(os.Stdout)
				break
			}
		}
	}

}

func (conf *Configuration) GetProgressbarDescriptionForVerb(verb string, repo Repository) string {
	trim := func(in string) string {
		return strings.TrimPrefix(filepath.ToSlash(in), conf.BaseDirectory+"/")
	}

	maxLength := len(fmt.Sprintf("%s %s", verb, trim(conf.Repositories.LongestName())))
	description := fmt.Sprintf("%s %s", verb, trim(repo.Directory))
	result := description + strings.Repeat(".", maxLength-len(description))

	return result
}

func (conf Configuration) GetToken() string {
	logger := loggerEntry

	host := util.GetHostnameFromPath(conf.BaseURL)
	logger.Debugf("Retrieving token for host: %s", host)

	token, _ := auth.TokenForHost(host)
	logger.Debugf("Retrieved token: %t", len(token) > 0)
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
	subDirectories := make(map[string]bool)
	for _, repo := range conf.Repositories {
		bar.Describe(util.CheckColors(color.RedString, conf.GetProgressbarDescriptionForVerb("Removing", repo)))
		subDirectories[filepath.Dir(repo.Directory)] = true
		util.FatalIfError(os.RemoveAll(repo.Directory))
		bar.Inc()
	}

	if conf.BaseDirectory != "." {
		util.FatalIfError(os.RemoveAll(conf.BaseDirectory))
	} else if conf.SubDirectories {
		for folder := range subDirectories {
			util.FatalIfError(os.Remove(folder))
		}
	}

	fmt.Println(util.CheckColors(color.GreenString, "Successfully removed repositories from local filesystem."))
}

func (conf Configuration) Save() {
	c, err := config.Read()
	util.FatalIfError(err)

	buffer := bytes.NewBuffer(nil)
	bar := newBinaryProgressbar().Describe(util.CheckColors(color.BlueString, "Saving..."))
	util.FatalIfError(yaml.NewEncoder(io.MultiWriter(buffer, bar)).Encode(conf))
	_ = bar.Clear()

	c.Set([]string{configKey}, buffer.String())
	util.FatalIfError(config.Write(c))

	fmt.Println(util.CheckColors(color.GreenString, "Configuration saved. You can now pull your repositories."))
}

func newBinaryProgressbar() *util.Progressbar {
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
