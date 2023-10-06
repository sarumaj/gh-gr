package configfile

import (
	"bufio"
	"bytes"
	"encoding/xml"
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
	ConfigInvalidFormat = "Invalid format %q. Supported formats are: [%s]."
	ConfigNotFound      = "No configuration found. Make sure to run 'init' to create initial configuration " +
		"or run 'import' to import configuration from stdin."
	ConfigShouldNotExist = "Configuration already exists. " +
		"Please run 'update' if you want to update your settings. " +
		"Alternatively, run 'remove' if you want to setup from scratch once again."
)

var urlRegex = regexp.MustCompile(`(?P<Schema>[^:]+://)(?P<Creds>[^@]+@)?(?P<Hostpath>.+)`)
var loggerEntry = util.Logger.WithFields(logrus.Fields{"mod": "configfile"})

// Configuration holds gr configuration data
type Configuration struct {
	XMLName        xml.Name      `json:"-" yaml:"-"`
	Username       string        `json:"username" yaml:"username"`
	Fullname       string        `json:"fullname" yaml:"fullname"`
	Email          string        `json:"email,omitempty" yaml:"email,omitempty"`
	BaseDirectory  string        `json:"baseDirectory" yaml:"baseDirectory"`
	BaseURL        string        `json:"baseURL" yaml:"baseURL"`
	Concurrency    uint          `json:"concurrency" yaml:"concurrency"`
	SubDirectories bool          `json:"subDirectories" yaml:"subDirectories"`
	Verbose        bool          `json:"verbose" yaml:"verbose"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`
	Excluded       []string      `json:"exluded,omitempty" yaml:"exluded,omitempty"`
	Included       []string      `json:"included,omitempty" yaml:"included,omitempty"`
	Repositories   Repositories  `json:"repositories" yaml:"repositories"`
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

func Import(format string) {
	enc, ok := supportedEncoders[format]
	if !ok {
		supportedEncoders := strings.Join(GetListOfSupportedFormats(true), ", ")
		fmt.Fprintln(os.Stderr, util.CheckColors(color.RedString, ConfigInvalidFormat, format, supportedEncoders))
		return
	}

	raw, err := io.ReadAll(os.Stdin)
	util.FatalIfError(err)

	var conf Configuration
	util.FatalIfError(enc.Decoder(bytes.NewReader(raw)).Decode(&conf))

	conf.Save()
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
		Included:       make([]string, len(conf.Included)),
		Excluded:       make([]string, len(conf.Excluded)),
		Repositories:   make(Repositories, len(conf.Repositories)),
	}

	_ = copy(n.Excluded, conf.Excluded)
	_ = copy(n.Included, conf.Included)
	_ = copy(n.Repositories, conf.Repositories)

	return n
}

func (conf Configuration) Display(format string, export bool) {
	reader, writer := io.Pipe()

	enc, ok := supportedEncoders[format]
	if !ok {
		supportedEncoders := strings.Join(GetListOfSupportedFormats(true), ", ")
		fmt.Fprintln(os.Stderr, util.CheckColors(color.RedString, ConfigInvalidFormat, format, supportedEncoders))
		return
	}

	go func() {
		defer writer.Close()
		util.FatalIfError(enc.Encoder(writer).Encode(conf))
	}()

	interactive := !export &&
		term.IsTerminal(os.Stdout) &&
		term.IsTerminal(os.Stdin) &&
		util.UseColors()

	for iter, noLines, scanner := 0, 10, bufio.NewScanner(reader); scanner.Scan(); iter++ {
		fmt.Fprintln(os.Stdout, scanner.Text())

		if interactive && iter > 0 && iter%noLines == 0 {
			fmt.Fprint(os.Stdout, color.BlueString("(more):"))

			var in string
			fmt.Fscanln(os.Stdin, &in)

			// move one line up and use carriage return to move to the beginning of line
			fmt.Fprint(os.Stdout, "\033[1A"+strings.Repeat(" ", len("(more):")+len(in))+"\r")

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
