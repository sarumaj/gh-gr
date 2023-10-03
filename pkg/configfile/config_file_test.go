package configfile

import (
	"os"
	"testing"

	monkey "bou.ke/monkey"
	auth "github.com/cli/go-gh/v2/pkg/auth"
	config "github.com/cli/go-gh/v2/pkg/config"
	"github.com/sarumaj/gh-gr/pkg/util"
)

const testConfiguration = `
git_protocol: https
editor:
prompt: enabled
pager:
http_unix_socket:
browser:
gr.conf: |
    username: user
    fullname: user
    email: 12345678-user@users.noreply.github.com
    baseDirectory: github
    baseURL: github.com
    concurrency: 16
    subDirectories: true
    verbose: false
    repositories:
    - url: https://github.com/user/repository.git
      directory: github/user/repository
      branch: main
`

func TestConfigurationAuthenticate(t *testing.T) {
	conf := &Configuration{Username: "user"}
	guard := monkey.Patch(auth.TokenForHost, func(string) (string, string) { return "token", "" })
	defer guard.Unpatch()

	for _, tt := range []struct {
		name string
		args string
		want string
	}{
		{"test#1", "http://example.com", "http://user:token@example.com"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args
			conf.Authenticate(&got)
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
				guard := monkey.Patch(config.Read, func() (*config.Config, error) { return config.ReadFromString(tt.args), nil })
				defer guard.Unpatch()
				*got = ConfigurationExists()
			}(&got)

			if got != tt.want {
				t.Errorf(`ConfigurationExists() failed: got: %t, want %t`, got, tt.want)
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
				guard := monkey.Patch(util.FatalIfError, func(err error) {
					if err != nil {
						capturedErr = err
					}
				})
				defer guard.Unpatch()

				guard = monkey.Patch(config.Read, func() (*config.Config, error) { return config.ReadFromString(tt.args), nil })
				defer guard.Unpatch()

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
				guard := monkey.Patch(util.FatalIfError, func(err error) {
					if err != nil {
						capturedErr = err
					}
				})
				defer guard.Unpatch()

				guard = monkey.Patch(config.Read, func() (*config.Config, error) { return config.ReadFromString(tt.args), nil })
				defer guard.Unpatch()

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
				guard := monkey.Patch(util.FatalIfError, func(err error) {
					if err != nil {
						capturedErr = err
					}
				})
				defer guard.Unpatch()

				guard = monkey.Patch(config.Read, func() (*config.Config, error) { return config.ReadFromString(tt.args), nil })
				defer guard.Unpatch()

				*conf = Load()
				(*conf).Save()
			}(&conf)

			if (capturedErr != nil && !tt.expectErr) || (capturedErr == nil && tt.expectErr) {
				t.Errorf(`Load() failed: %v`, capturedErr)
			}
		})
	}
}
