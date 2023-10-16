package configfile

import (
	"fmt"
	"io"

	color "github.com/fatih/color"
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
					e.SetColors(&jc.Colors{
						Null:   jc.Color(util.CSI + fmt.Sprintf("%d", color.FgHiMagenta) + util.CCT),
						Bool:   jc.Color(util.CSI + fmt.Sprintf("%d", color.FgHiMagenta) + util.CCT),
						Number: jc.Color(util.CSI + fmt.Sprintf("%d", color.FgHiMagenta) + util.CCT),
						String: jc.Color(util.CSI + fmt.Sprintf("%d", color.FgHiGreen) + util.CCT),
						Key:    jc.Color(util.CSI + fmt.Sprintf("%d", color.FgHiCyan) + util.CCT),
						Bytes:  jc.Color(util.CSI + fmt.Sprintf("%d", color.FgHiGreen) + util.CCT),
						Time:   jc.Color(util.CSI + fmt.Sprintf("%d", color.FgHiMagenta) + util.CCT),
					})
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

// Get list of supported formats.
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
