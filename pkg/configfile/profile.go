package configfile

import (
	"fmt"

	resources "github.com/sarumaj/gh-gr/pkg/restclient/resources"
	util "github.com/sarumaj/gh-gr/pkg/util"
)

// Profile holds the context of authenticated user profile.
type Profile struct {
	Username string `json:"username" yaml:"username"`
	Fullname string `json:"fullname" yaml:"fullname"`
	Email    string `json:"email,omitempty" yaml:"email,omitempty"`
	Host     string `json:"host" yaml:"host"`
}

// Create new profile for given GitHub user and host.
func NewProfile(user *resources.User, host string) *Profile {
	profile := &Profile{
		Username: user.Login,
		Fullname: user.Name,
		Email:    user.Email,
		Host:     host,
	}

	if profile.Host == "" {
		profile.Host = util.GetHostnameFromPath(user.URL)
	}

	if profile.Email == "" {
		profile.Email = fmt.Sprintf("%d-%s@users.noreply.github.com", user.ID, user.Login)
	}

	return profile
}

type Profiles []Profile

// Append profile (only if not present).
func (p *Profiles) Append(profile *Profile) {
	if profile != nil && !p.Has(*profile) {
		*p = append(*p, *profile)
	}
}

// Check if profile is enlisted (Username and Host are considered to be unique).
func (p Profiles) Has(profile Profile) bool {
	for _, own := range p {
		if own.Host == profile.Host && own.Username == profile.Username {
			return true
		}
	}

	return false
}

// Map profiles to hosts: <host> => <Profile>.
func (p Profiles) ToMap() map[string]Profile {
	m := make(map[string]Profile)
	for _, profile := range p {
		m[profile.Host] = profile
	}

	return m
}
