package configfile

import (
	"testing"

	monkey "bou.ke/monkey"
	auth "github.com/cli/go-gh/v2/pkg/auth"
)

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
