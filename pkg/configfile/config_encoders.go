package configfile

import (
	"encoding/json"
	"fmt"
	"io"

	yaml "gopkg.in/yaml.v3"
)

type (
	decoder     interface{ Decode(any) error }
	encoder     interface{ Encode(any) error }
	encoderPair struct {
		Encoder func(w io.Writer) encoder
		Decoder func(r io.Reader) decoder
	}
)

var supportedEncoders = map[string]encoderPair{
	"json": {
		Encoder: func(w io.Writer) encoder {
			e := json.NewEncoder(w)
			e.SetIndent("", "\t")
			return e
		},
		Decoder: func(r io.Reader) decoder { return json.NewDecoder(r) },
	},
	"yaml": {
		Encoder: func(w io.Writer) encoder { return yaml.NewEncoder(w) },
		Decoder: func(r io.Reader) decoder { return yaml.NewDecoder(r) },
	},
}

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
