package configfile

import (
	"fmt"

	"github.com/sarumaj/gh-gr/pkg/restclient/resources"
	"github.com/sarumaj/gh-gr/pkg/util"
)

type Profile struct {
	Username string `json:"username" yaml:"username"`
	Fullname string `json:"fullname" yaml:"fullname"`
	Email    string `json:"email,omitempty" yaml:"email,omitempty"`
	Host     string `json:"host" yaml:"host"`
}

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

func (p *Profiles) Append(profile *Profile) {
	*p = append(*p, *profile)
}

func (p Profiles) ToMap() map[string]Profile {
	m := make(map[string]Profile)
	for _, profile := range p {
		m[profile.Host] = profile
	}

	return m
}
