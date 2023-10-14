package configfile

import (
	"fmt"
	"io"

	json "github.com/goccy/go-json"
	yaml "github.com/goccy/go-yaml"
	jc "github.com/neilotoole/jsoncolor"
	util "github.com/sarumaj/gh-gr/pkg/util"
)

var supportedEncoders = func() map[string]encoderPair {
	return map[string]encoderPair{
		"json": {
			Encoder: func(w io.Writer, colored bool) encoder {
				if colored {
					e := jc.NewEncoder(w)
					e.SetColors(jc.DefaultColors())
					e.SetIndent("", "  ")
					return e
				}
				e := json.NewEncoder(w)
				e.SetIndent("", "  ")
				return e
			},
			Decoder: func(r io.Reader) decoder {
				e := json.NewDecoder(r)
				e.DisallowUnknownFields()
				return e
			},
		},
		"yaml": {
			Encoder: func(w io.Writer, colored bool) encoder {
				if colored {
					return util.NewColoredYAMLEncoder(w)
				}
				return yaml.NewEncoder(w)
			},
			Decoder: func(r io.Reader) decoder {
				return yaml.NewDecoder(r)
			},
		},
	}
}()

type (
	decoder     interface{ Decode(any) error }
	encoder     interface{ Encode(any) error }
	encoderPair struct {
		Encoder func(io.Writer, bool) encoder
		Decoder func(io.Reader) decoder
	}
)

func GetListOfSupportedFormats(quote bool) (formats []string) {
	for format := range supportedEncoders {
		if quote {
			formats = append(formats, fmt.Sprintf("%q", format))
		} else {
			formats = append(formats, format)
		}
	}

	return
}
