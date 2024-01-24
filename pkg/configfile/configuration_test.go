package configfile

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	monkey "bou.ke/monkey"
	auth "github.com/cli/go-gh/v2/pkg/auth"
	config "github.com/cli/go-gh/v2/pkg/config"
	resources "github.com/sarumaj/gh-gr/pkg/restclient/resources"
	util "github.com/sarumaj/gh-gr/pkg/util"
	supererrors "github.com/sarumaj/go-super/errors"
)

const testConfiguration = `
git_protocol: https
editor:
prompt: enabled
pager:
http_unix_socket:
browser:
gr.conf: |
    baseDirectory: github
    directoryPath: /
    profiles:
    - host: github.com
      username: user
      fullname: user
      email: 12345678-user@users.noreply.github.com
    concurrency: 16
    subDirectories: true
    verbose: false
    repositories:
    - URL: https://github.com/user/repository.git
      directory: github/user/repository
      branch: main
`

func TestConfigurationAuthenticateURL(t *testing.T) {
	conf := &Configuration{
		Profiles: Profiles{{Username: "user", Host: "example.com"}},
	}

	guard := monkey.Patch(util.PrintlnAndExit, func(format string, a ...any) { _, _ = fmt.Printf(format, a...) })
	defer guard.Unpatch()

	guard = monkey.Patch(auth.TokenForHost, func(string) (string, string) { return "token", "" })
	defer guard.Unpatch()

	guard = monkey.Patch(auth.KnownHosts, func() []string { return []string{"example.com"} })
	defer guard.Unpatch()

	for _, tt := range []struct {
		name string
		args string
		want string
	}{
		{"test#1", "http://example.com", "http://user:token@example.com"},
		{"test#2", "http://invalid:invalid@example.com", "http://user:token@example.com"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args
			conf.AuthenticateURL(&got)
			if got != tt.want {
				t.Errorf(`conf.Authenticate(&(%q)) failed: got: %q, want %q`, tt.args, got, tt.want)
			}
		})
	}
}

func TestConfigurationExists(t *testing.T) {
	for _, tt := range []struct {
		name string
		args string
		want bool
	}{
		{"test#1", "", false},
		{"test#2", testConfiguration, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			func(got *bool) {
				backup := configReader
				defer func() { configReader = backup }()
				configReader = func() (*config.Config, error) { return config.ReadFromString(tt.args), nil }
				*got = ConfigurationExists()
			}(&got)

			if got != tt.want {
				t.Errorf(`ConfigurationExists() failed: got: %t, want %t`, got, tt.want)
			}
		})
	}
}

func TestConfigurationFilterRepositories(t *testing.T) {
	type args struct {
		conf  *Configuration
		repos []resources.Repository
	}

	makeRepo := func(name string) resources.Repository {
		return resources.Repository{
			FullName: name,
			Permissions: resources.Permissions{
				Pull: true,
				Push: true,
			},
		}
	}

	makeRepos := func(names ...string) (repos []resources.Repository) {
		for _, name := range names {
			repos = append(repos, makeRepo(name))
		}

		return
	}

	for _, tt := range []struct {
		name string
		args args
		want []resources.Repository
	}{
		{"test#1",
			args{&Configuration{Included: []string{}, Excluded: []string{}}, makeRepos("org1/repo1", "org2/repo1", "org2/repo2")},
			makeRepos("org1/repo1", "org2/repo1", "org2/repo2")},
		{"test#2",
			args{&Configuration{Included: []string{"org2/.*"}, Excluded: []string{}}, makeRepos("org1/repo1", "org2/repo1", "org2/repo2")},
			makeRepos("org2/repo1", "org2/repo2")},
		{"test#3",
			args{&Configuration{Included: []string{"org2/.*"}, Excluded: []string{"org2/repo2"}}, makeRepos("org1/repo1", "org2/repo1", "org2/repo2")},
			makeRepos("org2/repo1")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.repos
			tt.args.conf.FilterRepositories(&got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("(%v).FilterRepositories(%p) failed: got: %d, want: %d", tt.args.conf, &got, len(got), len(tt.want))
			}
		})
	}
}

func TestConfigurationGeneralize(t *testing.T) {
	conf := &Configuration{}
	for _, tt := range []struct {
		name string
		args string
		want string
	}{
		{"test#1", "", ""},
		{"test#2", "https://user:pass@example.com", "https://example.com"},
		{"test#3", "https://example.com", "https://example.com"},
		{"test#4", "https://example.com?q=1", "https://example.com?q=1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args
			conf.GeneralizeURL(&got)
			if got != tt.want {
				t.Errorf(`conf.Generalize(&(%q)) failed: got: %q, want %q`, tt.args, got, tt.want)
			}
		})
	}
}

func TestConfigurationLoad(t *testing.T) {
	for _, tt := range []struct {
		name      string
		args      string
		expectErr bool
	}{
		{"test#1", "", true},
		{"test#2", testConfiguration, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var capturedErr error
			func() {
				guard := monkey.Patch(supererrors.Except, func(err error, ignore ...error) {
					if err != nil {
						for _, e := range ignore {
							if errors.Is(err, e) {
								return
							}
						}

						capturedErr = err
					}
				})
				defer guard.Unpatch()

				backup := configReader
				defer func() { configReader = backup }()
				configReader = func() (*config.Config, error) { return config.ReadFromString(tt.args), nil }

				_ = Load()
			}()

			if capturedErr != nil && !tt.expectErr {
				t.Errorf(`Load() failed: %v`, capturedErr)
			}
		})
	}
}

func TestConfigurationRemove(t *testing.T) {
	guard := monkey.Patch(os.RemoveAll, func(string) error { return nil })
	defer guard.Unpatch()

	guard = monkey.Patch(config.Write, func(*config.Config) error { return nil })
	defer guard.Unpatch()

	for _, tt := range []struct {
		name      string
		args      string
		expectErr bool
	}{
		{"test#1", "", true},
		{"test#2", testConfiguration, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var capturedErr error
			var conf *Configuration
			func(conf **Configuration) {
				guard := monkey.Patch(supererrors.Except, func(err error, ignore ...error) {
					if err != nil {
						for _, e := range ignore {
							if errors.Is(err, e) {
								return
							}
						}

						capturedErr = err
					}
				})
				defer guard.Unpatch()

				backup := configReader
				defer func() { configReader = backup }()
				configReader = func() (*config.Config, error) { return config.ReadFromString(tt.args), nil }

				*conf = Load()

				(*conf).Remove(true)
			}(&conf)

			if (capturedErr != nil && !tt.expectErr) || (capturedErr == nil && tt.expectErr) {
				t.Errorf(`Load() failed: %v`, capturedErr)
			}
		})
	}
}

func TestConfigurationSave(t *testing.T) {
	guard := monkey.Patch(config.Write, func(*config.Config) error { return nil })
	defer guard.Unpatch()

	for _, tt := range []struct {
		name      string
		args      string
		expectErr bool
	}{
		{"test#1", "", true},
		{"test#2", testConfiguration, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var capturedErr error
			var conf *Configuration
			func(conf **Configuration) {
				guard := monkey.Patch(supererrors.Except, func(err error, ignore ...error) {
					if err != nil {
						for _, e := range ignore {
							if errors.Is(err, e) {
								return
							}
						}

						capturedErr = err
					}
				})
				defer guard.Unpatch()

				backup := configReader
				defer func() { configReader = backup }()
				configReader = func() (*config.Config, error) { return config.ReadFromString(tt.args), nil }

				*conf = Load()
				(*conf).Save()
			}(&conf)

			if (capturedErr != nil && !tt.expectErr) || (capturedErr == nil && tt.expectErr) {
				t.Errorf(`Load() failed: %v`, capturedErr)
			}
		})
	}
}
