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
	"slices"
	"strings"
	"time"

	terminal "github.com/AlecAivazis/survey/v2/terminal"
	config "github.com/cli/go-gh/v2/pkg/config"
	prompter "github.com/cli/go-gh/v2/pkg/prompter"
	color "github.com/fatih/color"
	resources "github.com/sarumaj/gh-gr/v2/pkg/restclient/resources"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	supererrors "github.com/sarumaj/go-super/errors"
	logrus "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

// Attribute name to store the configuration inside of the GitHub CLI config.
const configKey = "gr.conf"

// Message, when authentication fails.
const AuthenticationFailed = "Authentication for %q failed. Make sure to configure GitHub CLI for %q."

// Message, when unsupported format for config provided.
const ConfigInvalidFormat = "Invalid format %q. Supported formats are: [%s]."

// Message, when config not found.
const ConfigNotFound = "No configuration found. Make sure to run 'init' to create initial configuration " +
	"or run 'import' to import configuration from stdin."

// Message, when config exists, although it should not.
const ConfigShouldNotExist = "Configuration already exists. " +
	"Please run 'update' if you want to update your settings. " +
	"Alternatively, run 'remove' if you want to setup from scratch once again."

// Default destination for export.
const DefaultExportDestination = "stdout"

// Default source for import.
const DefaultImportSource = "stdin"

// Regular expression used to split URL into components.
var urlRegex = regexp.MustCompile(`(?P<Schema>[^:]+://)(?P<Creds>[^@]+@)?(?P<Hostpath>.+)`)

// Module logger.
var loggerEntry = util.Logger.WithFields(logrus.Fields{"mod": "configfile"})

// Prompter for interactive shell.
var prompt = func() *prompter.Prompter {
	c := util.Console()
	return prompter.New(c.Stdin(), c.Stdout(), c.Stderr())
}()

// Related to https://github.com/sarumaj/gh-gr/v2/issues/62
var configReader = func() func() (*config.Config, error) {
	switch read := any(config.Read).(type) {
	case func() (*config.Config, error):
		return read

	case func(*config.Config) (*config.Config, error):
		return func() (*config.Config, error) { return read(nil) }

	default:
		return func() (*config.Config, error) { return nil, fmt.Errorf(`config.Read is of unsupported type: %T`, read) }

	}
}()

// Configuration holds gr configuration data
type Configuration struct {
	BaseDirectory         string        `json:"baseDirectory" yaml:"baseDirectory"`
	AbsoluteDirectoryPath string        `json:"directoryPath" yaml:"directoryPath"`
	Profiles              Profiles      `json:"profiles" yaml:"profiles"`
	Concurrency           uint          `json:"concurrency" yaml:"concurrency"`
	SubDirectories        bool          `json:"subDirectories" yaml:"subDirectories"`
	SizeLimit             uint64        `json:"sizeLimit" yaml:"sizeLimit"`
	Timeout               time.Duration `json:"timeout" yaml:"timeout"`
	Excluded              []string      `json:"exluded,omitempty" yaml:"exluded,omitempty"`
	Included              []string      `json:"included,omitempty" yaml:"included,omitempty"`
	Total                 int64         `json:"total" yaml:"total"`
	Repositories          Repositories  `json:"repositories" yaml:"repositories"`
}

// AppendRepositories appends multiple repositories to the configuration and sorts them alphabetically by Directory.
func (conf *Configuration) AppendRepositories(user *resources.User, repos ...resources.Repository) {
	for _, repo := range repos {
		dir := repo.FullName
		if !conf.SubDirectories {
			dir = strings.ReplaceAll(dir, "/", "_")
			dir = strings.Replace(dir, user.Login+"_", "", 1)
		}

		dir = filepath.Join(conf.BaseDirectory, filepath.FromSlash(dir))
		util.PathSanitize(&dir)

		loggerEntry.Debugf("Appending %s", dir)

		conf.Repositories.Append(Repository{
			Branch:    repo.DefaultBranch,
			Directory: dir,
			ParentURL: repo.Parent.CloneURL,
			Public:    !repo.Private,
			Size:      util.IntToSizeBytes(repo.Size, 1024, 3),
			URL:       repo.CloneURL,
		})
	}

	slices.SortFunc(conf.Repositories, func(a, b Repository) int {
		switch {
		case a.Directory > b.Directory:
			return 1

		case a.Directory < b.Directory:
			return -1

		default:
			return 0
		}
	})

	conf.Total = int64(len(conf.Repositories))
	loggerEntry.Debugf("Configured %d repositories", conf.Total)
}

// AuthenticateURL encodes username and token into URL.
// In the case, no matching token can be found for given URL, emit message and exit.
func (conf Configuration) AuthenticateURL(targetURL *string) {
	loggerEntry.Debugf("Authenticating URL: %v", targetURL)
	if targetURL == nil || *targetURL == "" || !urlRegex.MatchString(*targetURL) {
		loggerEntry.Debugf("Got empty or invalid URL: %#v", targetURL)
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
				loggerEntry.Debugf("Failed to authenticate: %s, %v", *targetURL, err)
				continue
			}

			loggerEntry.Debugf("Authenticated: %s", *targetURL)
			*targetURL = parsed.String()

			return
		}
	}

	util.PrintlnAndExit(util.Console().CheckColors(color.RedString, AuthenticationFailed, *targetURL, hostname))
}

// Remove local repositories which are not enlisted.
func (conf Configuration) Cleanup() {
	c := util.Console()
	interactive := c.IsTerminal(true, true, true)

	untracked := conf.ListUntracked()
	if len(untracked) == 0 {
		_ = supererrors.ExceptFn(supererrors.W(
			fmt.Fprintln(c.Stdout(), c.CheckColors(color.GreenString, "No untracked directories to remove.")),
		))
		return
	}

	var selected []int
	if interactive {
		selected = supererrors.ExceptFn(supererrors.W(
			prompt.MultiSelect(
				c.CheckColors(
					color.RedString,
					"DANGER!!! ",
				)+"Select untracked repositories to remove:",
				untracked,
				untracked,
			),
		), terminal.InterruptErr)

		if supererrors.LastErrorWas(terminal.InterruptErr) {
			os.Exit(0)
		}
	}
	if len(selected) == 0 {
		return
	}

	if interactive {
		if !supererrors.ExceptFn(supererrors.W(
			prompt.Confirm(
				c.CheckColors(
					color.RedString,
					"DANGER!!! ",
				)+fmt.Sprintf("You will delete %d untracked local directories! Are you sure?", len(selected)),
				false,
			),
		), terminal.InterruptErr) {

			return
		}

		if supererrors.LastErrorWas(terminal.InterruptErr) {
			os.Exit(0)
		}
	}

	defer util.Chdir(conf.AbsoluteDirectoryPath).Popd()

	var tmp Configuration
	for _, index := range selected {
		tmp.Repositories = append(tmp.Repositories, Repository{Directory: untracked[index]})
	}

	bar := util.NewProgressbar(len(tmp.Repositories))
	for _, repo := range tmp.Repositories {
		_ = bar.Describe(c.CheckColors(color.RedString, tmp.GetProgressbarDescriptionForVerb("Removing", repo)))
		supererrors.Except(os.RemoveAll(repo.Directory), os.ErrNotExist)
		_ = bar.Inc()
	}

	_ = supererrors.ExceptFn(supererrors.W(
		fmt.Fprintln(c.Stdout(), c.CheckColors(color.GreenString, "Successfully removed %d untracked directories from local filesystem.", len(selected))),
	))
}

// Clone config.
func (conf *Configuration) Copy() *Configuration {
	n := &Configuration{
		BaseDirectory:         conf.BaseDirectory,
		AbsoluteDirectoryPath: conf.AbsoluteDirectoryPath,
		Profiles:              make(Profiles, len(conf.Profiles)),
		Concurrency:           conf.Concurrency,
		SubDirectories:        conf.SubDirectories,
		SizeLimit:             conf.SizeLimit,
		Timeout:               conf.Timeout,
		Included:              make([]string, len(conf.Included)),
		Excluded:              make([]string, len(conf.Excluded)),
		Repositories:          make(Repositories, len(conf.Repositories)),
		Total:                 conf.Total,
	}

	_ = copy(n.Excluded, conf.Excluded)
	_ = copy(n.Included, conf.Included)
	_ = copy(n.Profiles, conf.Profiles)
	_ = copy(n.Repositories, conf.Repositories)

	return n
}

// Display flushes config into Stdout.
// Supports multiple formats and partial emission (if !export).
func (conf Configuration) Display(format, output string, export bool, filters []string) {
	reader, writer := io.Pipe()
	c := util.Console()

	enc, ok := supportedEncoders[format]
	if !ok {
		supportedEncoders := strings.Join(GetListOfSupportedFormats(true), ", ")
		util.PrintlnAndExit(c.CheckColors(color.RedString, ConfigInvalidFormat, format, supportedEncoders))
	}

	if export { // prevent flood of logs into stdout
		out := util.Logger.Out
		util.Logger.SetOutput(io.Discard)
		defer func(w io.Writer) { util.Logger.SetOutput(w) }(out)
	}

	if len(filters) > 0 {
		var repositories []Repository
		for _, repository := range conf.Repositories {
			if util.PatternList(filters).GlobMatch(util.StripPathPrefix(repository.Directory, 1)) {
				repositories = append(repositories, repository)
			}
		}
		conf.Repositories = repositories
	}

	go func() {
		defer writer.Close()
		supererrors.Except(enc.Encoder(writer, !export && c.ColorsEnabled()).Encode(conf))
	}()

	var target io.Writer = c.Stdout()
	if output != DefaultExportDestination && export {
		target = supererrors.ExceptFn(supererrors.W(os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)))
	}

	isTerminal := c.IsTerminal(true, false, true)
	for iter, noLines, scanner := 0, 10, bufio.NewScanner(reader); scanner.Scan(); iter++ {
		_ = supererrors.ExceptFn(supererrors.W(fmt.Fprintln(target, scanner.Text())))

		if !export && isTerminal && iter > 0 && iter%noLines == 0 {
			_ = supererrors.ExceptFn(supererrors.W(
				fmt.Fprint(target, c.CheckColors(color.BlueString, "(more:)")),
			))

			var in string
			_, err := fmt.Fscanln(c.Stdin(), &in)

			// move one line up and use carriage return to move to the beginning of line
			_ = supererrors.ExceptFn(supererrors.W(
				fmt.Fprint(target, util.MUP+strings.Repeat(" ", len("(more):")+len(in))+"\r"),
			))

			if err != nil {
				continue
			}

			switch strings.ToLower(in) {

			case "exit", "quit", "q":
				_ = supererrors.ExceptFn(supererrors.W(fmt.Fprintln(target)))
				return

			}
		}
	}
}

// Filter repositories.
// Applied filters: included || !excluded || size below x || archived || disabled || !pullPermission || !pushPermission.
func (conf *Configuration) FilterRepositories(repositories *[]resources.Repository) {
	for index, total := 0, len(*repositories); index < total; index++ {
		switch repo := (*repositories)[index]; {

		case
			// not explicitly included
			len(conf.Included) > 0 && !util.PatternList(conf.Included).RegexMatch(repo.FullName, conf.Timeout),
			// explicitly excluded
			len(conf.Excluded) > 0 && util.PatternList(conf.Excluded).RegexMatch(repo.FullName, conf.Timeout),
			// repository size exceeds size limit
			conf.SizeLimit > 0 && uint64(repo.Size) > conf.SizeLimit,
			// repository is archived or disabled
			repo.Archived || repo.Disabled,
			// lacking pull and push permissions
			!repo.Permissions.Pull || !repo.Permissions.Push:

			loggerEntry.Debugf("Skipping %s", repo.FullName)
			// remove the repository at index
			*repositories = append((*repositories)[:index], (*repositories)[index+1:]...)[: total-1 : total-1]
			// move index back to point at the next repository which now occupies the position of the removed one
			index, total = index-1, total-1

		}
	}
}

// GeneralizeURL removes username and token from URL.
func (conf Configuration) GeneralizeURL(targetURL *string) {
	if targetURL == nil || *targetURL == "" || !urlRegex.MatchString(*targetURL) {
		loggerEntry.Debugf("Got empty or invalid URL: %#v", targetURL)
		return
	}

	loggerEntry.Debugf("Generalizing URL: %s", *targetURL)

	*targetURL = urlRegex.ReplaceAllString(*targetURL, "${Schema}${Hostpath}")
	loggerEntry.Debugf("Generalized: %s", *targetURL)
}

// Produce progressbar description considering the length of the repository with the longest name.
func (conf *Configuration) GetProgressbarDescriptionForVerb(verb string, repo Repository) string {
	trim := func(in string) string {
		util.PathSanitize(&in, &conf.BaseDirectory)
		return strings.TrimPrefix(in, conf.BaseDirectory+"/")
	}

	maxLength := len(fmt.Sprintf("%s %s...", verb, trim(conf.Repositories.LongestName())))
	description := fmt.Sprintf("%s %s...", verb, trim(repo.Directory))
	result := description + strings.Repeat(".", maxLength-len(description))

	return result
}

// List local repositories which are not enlisted any longer.
func (conf Configuration) ListUntracked() (untracked []string) {
	util.PathSanitize(&conf.BaseDirectory)
	defer util.Chdir(conf.AbsoluteDirectoryPath).Popd()

	files := supererrors.ExceptFn(supererrors.W(filepath.Glob(conf.BaseDirectory + "/*")))
	if conf.SubDirectories {
		parents := supererrors.ExceptFn(supererrors.W(filepath.Glob(conf.BaseDirectory + "/*/*")))
		files = append(files, parents...)
	}

	for _, f := range files {
		if !isRepoDir(f, conf.Repositories) {
			untracked = append(untracked, f)
		}
	}

	return
}

// Remove config.
// If purge, remove all local repositories.
func (conf Configuration) Remove(purge bool) {
	ghconf := supererrors.ExceptFn(supererrors.W(configReader()))

	supererrors.Except(ghconf.Remove([]string{configKey}))
	supererrors.Except(config.Write(ghconf))

	c := util.Console()
	_ = supererrors.ExceptFn(supererrors.W(
		fmt.Fprintln(c.Stdout(), c.CheckColors(color.GreenString, "Configuration removed.")),
	))

	if !purge {
		return
	}

	if c.IsTerminal(true, true, true) {
		if !supererrors.ExceptFn(supererrors.W(
			prompt.Confirm(
				c.CheckColors(
					color.RedString,
					"DANGER!!! ",
				)+"You will delete all local repositories! Are you sure?",
				false,
			),
		), terminal.InterruptErr) {

			return
		}

		if supererrors.LastErrorWas(terminal.InterruptErr) {
			os.Exit(0)
		}
	}

	defer util.Chdir(conf.AbsoluteDirectoryPath).Popd()

	bar := util.NewProgressbar(len(conf.Repositories))
	subDirectories := make(map[string]bool)
	for _, repo := range conf.Repositories {
		_ = bar.Describe(c.CheckColors(color.RedString, conf.GetProgressbarDescriptionForVerb("Removing", repo)))
		supererrors.Except(os.RemoveAll(repo.Directory), os.ErrNotExist)
		_ = bar.Inc()

		if conf.SubDirectories {
			subDirectories[filepath.Dir(repo.Directory)] = true
		}
	}

	if conf.BaseDirectory != "." {
		supererrors.Except(os.RemoveAll(conf.BaseDirectory), os.ErrNotExist)

	} else if conf.SubDirectories {
		for folder := range subDirectories {
			supererrors.Except(os.Remove(folder), os.ErrNotExist)
		}

	}
	_ = supererrors.ExceptFn(supererrors.W(
		fmt.Fprintln(c.Stdout(), c.CheckColors(color.GreenString, "Successfully removed repositories from local filesystem.")),
	))
}

// Transform base directory into UNIx style path and set absolute directory path.
func (conf *Configuration) SanitizeDirectory() {
	if filepath.IsAbs(conf.BaseDirectory) {
		conf.AbsoluteDirectoryPath = filepath.Dir(conf.BaseDirectory)
		conf.BaseDirectory = filepath.Base(conf.BaseDirectory)

	} else {
		conf.AbsoluteDirectoryPath = filepath.Dir(supererrors.ExceptFn(supererrors.W(filepath.Abs(conf.BaseDirectory))))

	}

	util.PathSanitize(&conf.BaseDirectory, &conf.AbsoluteDirectoryPath)
}

// Save configuration into GitHub CLI config.
func (conf Configuration) Save() {
	ghconf := supererrors.ExceptFn(supererrors.W(configReader()))

	c := util.Console()
	buffer := bytes.NewBuffer(nil)
	bar := newBinaryProgressbar().Describe(c.CheckColors(color.BlueString, "Saving..."))
	supererrors.Except(yaml.NewEncoder(io.MultiWriter(buffer, bar)).Encode(conf))
	_ = bar.Clear()

	ghconf.Set([]string{configKey}, buffer.String())
	supererrors.Except(config.Write(ghconf))

	_ = supererrors.ExceptFn(supererrors.W(
		fmt.Fprintln(c.Stdout(), c.CheckColors(color.GreenString, "Configuration saved. You can now pull %d repositories.", conf.Total)),
	))
}

// Check if configuration exists within GitHub CLI config.
func ConfigurationExists() bool {
	c, err := configReader()
	if err != nil {
		return false
	}

	raw, err := c.Get([]string{configKey})

	return err == nil && len(raw) > 0
}

// Load configuration from GitHub CLI config.
func Load() *Configuration {
	ghconf := supererrors.ExceptFn(supererrors.W(configReader()))
	content := supererrors.ExceptFn(supererrors.W(ghconf.Get([]string{configKey})))

	var conf Configuration
	c := util.Console()
	bar := newBinaryProgressbar().Describe(c.CheckColors(color.BlueString, "Loading..."))
	supererrors.Except(yaml.NewDecoder(io.TeeReader(strings.NewReader(content), bar)).Decode(&conf))
	_ = bar.Clear()

	return &conf
}

// Import configuration from Stdin or local file.
func Import(format, input string) {
	c := util.Console()

	var reader io.Reader
	if input != DefaultImportSource && util.PathExists(input) {
		reader = bufio.NewReader(supererrors.ExceptFn(supererrors.W(os.OpenFile(input, os.O_RDONLY, os.ModePerm))))

	} else if c.IsTerminal(true, true, true) {
		if fileList := util.ListFilesByExtension("."+format, 2); len(fileList) > 0 {
			input = fileList[supererrors.ExceptFn(supererrors.W(
				prompt.Select(
					"Select file to import the configuration from:",
					fileList[0],
					fileList,
				),
			), terminal.InterruptErr)]

		} else {
			input = supererrors.ExceptFn(supererrors.W(prompt.Input("Provide path to configuration file:", "")))

		}

		if supererrors.LastErrorWas(terminal.InterruptErr) {
			os.Exit(0)
		}

		if !supererrors.ExceptFn(supererrors.W(
			prompt.Confirm(
				c.CheckColors(
					color.RedString,
					"DANGER!!! ",
				)+"You will overwrite the configuration! Are you sure?",
				false,
			),
		), terminal.InterruptErr) {

			return
		}

		if supererrors.LastErrorWas(terminal.InterruptErr) {
			os.Exit(0)
		}

		reader = bufio.NewReader(supererrors.ExceptFn(supererrors.W(os.OpenFile(input, os.O_RDONLY, os.ModePerm))))
	} else {
		reader = bufio.NewReader(c.Stdin())
	}

	enc, ok := supportedEncoders[format]
	if !ok {
		supportedEncoders := strings.Join(GetListOfSupportedFormats(true), ", ")
		util.PrintlnAndExit(c.CheckColors(color.RedString, ConfigInvalidFormat, format, supportedEncoders))
	}

	bar := newBinaryProgressbar().Describe(c.CheckColors(color.BlueString, "Importing..."))
	raw := supererrors.ExceptFn(supererrors.W(io.ReadAll(io.TeeReader(reader, bar))))

	var conf Configuration
	supererrors.Except(enc.Decoder(bytes.NewReader(raw)).Decode(&conf))

	conf.Save()
}
