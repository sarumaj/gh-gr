package configfile

type Profile struct {
	Username string `json:"username" yaml:"username"`
	Fullname string `json:"fullname" yaml:"fullname"`
	Email    string `json:"email,omitempty" yaml:"email,omitempty"`
	Host     string `json:"host" yaml:"host"`
}

type Profiles []Profile

func (p Profiles) ToMap() map[string]Profile {
	m := make(map[string]Profile, len(p))
	for _, profile := range p {
		m[profile.Host] = profile
	}

	return m
}
