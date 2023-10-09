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

	config "github.com/cli/go-gh/v2/pkg/config"
	prompter "github.com/cli/go-gh/v2/pkg/prompter"
	color "github.com/fatih/color"
	util "github.com/sarumaj/gh-gr/pkg/util"
	logrus "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const configKey = "gr.conf"

const (
	AuthenticationFailed = "Authentication for %q failed. Make sure to configure GitHub CLI for %q."
	ConfigInvalidFormat  = "Invalid format %q. Supported formats are: [%s]."
	ConfigNotFound       = "No configuration found. Make sure to run 'init' to create initial configuration " +
		"or run 'import' to import configuration from stdin."
	ConfigShouldNotExist = "Configuration already exists. " +
		"Please run 'update' if you want to update your settings. " +
		"Alternatively, run 'remove' if you want to setup from scratch once again."
)

var urlRegex = regexp.MustCompile(`(?P<Schema>[^:]+://)(?P<Creds>[^@]+@)?(?P<Hostpath>.+)`)
var loggerEntry = util.Logger.WithFields(logrus.Fields{"mod": "configfile"})

var prompt = prompter.New(util.Stdin(), util.Stdout(), util.Stderr())

// Configuration holds gr configuration data
type Configuration struct {
	BaseDirectory         string        `json:"baseDirectory" yaml:"baseDirectory"`
	AbsoluteDirectoryPath string        `json:"directoryPath" yaml:"directoryPath"`
	Profiles              Profiles      `json:"profiles" yaml:"profiles"`
	Concurrency           uint          `json:"concurrency" yaml:"concurrency"`
	SubDirectories        bool          `json:"subDirectories" yaml:"subDirectories"`
	Verbose               bool          `json:"verbose" yaml:"verbose"`
	Timeout               time.Duration `json:"timeout" yaml:"timeout"`
	Excluded              []string      `json:"exluded,omitempty" yaml:"exluded,omitempty"`
	Included              []string      `json:"included,omitempty" yaml:"included,omitempty"`
	Repositories          Repositories  `json:"repositories" yaml:"repositories"`
}

func (conf Configuration) Authenticate(targetURL *string) {
	loggerEntry.Debugf("Authenticating URL: %v", targetURL)
	if targetURL == nil || *targetURL == "" || !urlRegex.MatchString(*targetURL) {
		return
	}

	hostname := util.GetHostnameFromPath(*targetURL)
	profiles := conf.Profiles.ToMap()
	tokens := GetTokens()

	for host, token := range tokens {
		if profile, ok := profiles[host]; ok && hostname == host {
			parsed, err := url.Parse(urlRegex.ReplaceAllString(
				*targetURL,
				fmt.Sprintf("${Schema}%s:%s@${Hostpath}", profile.Username, token),
			))
			if err != nil {
				return
			}

			loggerEntry.Debugf("Authenticated: %s", *targetURL)
			*targetURL = parsed.String()

			return
		}
	}

	util.PrintlnAndExit(util.CheckColors(color.RedString, AuthenticationFailed, *targetURL, hostname))
}

func (conf *Configuration) Copy() *Configuration {
	n := &Configuration{
		BaseDirectory:         conf.BaseDirectory,
		AbsoluteDirectoryPath: conf.AbsoluteDirectoryPath,
		Profiles:              make(Profiles, len(conf.Profiles)),
		Concurrency:           conf.Concurrency,
		SubDirectories:        conf.SubDirectories,
		Verbose:               conf.Verbose,
		Timeout:               conf.Timeout,
		Included:              make([]string, len(conf.Included)),
		Excluded:              make([]string, len(conf.Excluded)),
		Repositories:          make(Repositories, len(conf.Repositories)),
	}

	_ = copy(n.Excluded, conf.Excluded)
	_ = copy(n.Included, conf.Included)
	_ = copy(n.Profiles, conf.Profiles)
	_ = copy(n.Repositories, conf.Repositories)

	return n
}

func (conf Configuration) Display(format string, export bool) {
	reader, writer := io.Pipe()
	defer writer.Close()

	enc, ok := supportedEncoders[format]
	if !ok {
		supportedEncoders := strings.Join(GetListOfSupportedFormats(true), ", ")
		util.PrintlnAndExit(util.CheckColors(color.RedString, ConfigInvalidFormat, format, supportedEncoders))
	}

	go func() {
		defer writer.Close()
		util.FatalIfError(enc.Encoder(writer).Encode(conf))
	}()

	interactive := !export && util.IsTerminal(true, false, true)

	for iter, noLines, scanner := 0, 10, bufio.NewScanner(reader); scanner.Scan(); iter++ {
		fmt.Fprintln(util.Stdout(), scanner.Text())

		if interactive && iter > 0 && iter%noLines == 0 {
			fmt.Fprint(util.Stdout(), color.BlueString("(more):"))

			var in string
			fmt.Fscanln(util.Stdin(), &in)

			// move one line up and use carriage return to move to the beginning of line
			fmt.Fprint(util.Stdout(), "\033[1A"+strings.Repeat(" ", len("(more):")+len(in))+"\r")

			if strings.HasPrefix(strings.ToLower(in), "q") {
				fmt.Fprintln(util.Stdout())
				break
			}
		}
	}
}

func (conf *Configuration) GetProgressbarDescriptionForVerb(verb string, repo Repository) string {
	trim := func(in string) string {
		util.PathSanitize(&in, &conf.BaseDirectory)
		return strings.TrimPrefix(in, conf.BaseDirectory+"/")
	}

	maxLength := len(fmt.Sprintf("%s %s", verb, trim(conf.Repositories.LongestName())))
	description := fmt.Sprintf("%s %s", verb, trim(repo.Directory))
	result := description + strings.Repeat(".", maxLength-len(description))

	return result
}

func (conf Configuration) Remove(purge bool) {
	c, err := config.Read()
	util.FatalIfError(err)

	util.FatalIfError(c.Remove([]string{configKey}))
	util.FatalIfError(config.Write(c))

	_, _ = fmt.Fprintln(util.Stdout(), util.CheckColors(color.GreenString, "Configuration removed."))

	if !purge {
		return
	}

	if util.IsTerminal(true, true, true) {
		confirm, err := prompt.Confirm(
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

	defer util.MoveToPath(conf.AbsoluteDirectoryPath)()

	bar := util.NewProgressbar(len(conf.Repositories))
	subDirectories := make(map[string]bool)
	for _, repo := range conf.Repositories {
		_ = bar.Describe(util.CheckColors(color.RedString, conf.GetProgressbarDescriptionForVerb("Removing", repo)))
		util.FatalIfError(os.RemoveAll(repo.Directory))
		_ = bar.Inc()

		if conf.SubDirectories {
			subDirectories[filepath.Dir(repo.Directory)] = true
		}
	}

	if conf.BaseDirectory != "." {
		util.FatalIfError(os.RemoveAll(conf.BaseDirectory))

	} else if conf.SubDirectories {
		for folder := range subDirectories {
			util.FatalIfError(os.Remove(folder))
		}

	}

	_, _ = fmt.Fprintln(util.Stdout(), util.CheckColors(color.GreenString, "Successfully removed repositories from local filesystem."))
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

	_, _ = fmt.Fprintln(util.Stdout(), util.CheckColors(color.GreenString, "Configuration saved. You can now pull your repositories."))
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
		util.PrintlnAndExit(util.CheckColors(color.RedString, ConfigInvalidFormat, format, supportedEncoders))
	}

	bar := newBinaryProgressbar().Describe(util.CheckColors(color.BlueString, "Importing..."))
	raw, err := io.ReadAll(io.TeeReader(util.Stdin(), bar))
	util.FatalIfError(err)

	var conf Configuration
	util.FatalIfError(enc.Decoder(bytes.NewReader(raw)).Decode(&conf))

	conf.Save()
}
