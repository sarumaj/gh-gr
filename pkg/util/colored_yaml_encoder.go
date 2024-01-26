package util

import (
	"bytes"
	"fmt"
	"io"
	"regexp"

	color "github.com/fatih/color"
	yaml "github.com/goccy/go-yaml"
	lexer "github.com/goccy/go-yaml/lexer"
	printer "github.com/goccy/go-yaml/printer"
)

// colorSequenceRegex is a regular expression for matching ANSI code sequences for color codes.
var colorSequenceRegex = regexp.MustCompile(regexp.QuoteMeta(CSI) + fmt.Sprintf("[^%[1]s]+%[1]s", CCT))

// coloredYAMLEncoder is a custom encoder to produce colored YAML output.
type coloredYAMLEncoder struct {
	b *bytes.Buffer
	w io.Writer
	*yaml.Encoder
}

// Encode encodes YAML and colorizes output.
func (enc *coloredYAMLEncoder) Encode(v any) error {
	if err := enc.Encoder.Encode(v); err != nil {
		return err
	}

	marshaled := enc.b.String()
	enc.b.Reset()

	p := &printer.Printer{
		LineNumber: true,
		LineNumberFormat: func(num int) string {
			return color.New(color.Bold).Sprintf("%2d | ", num)
		},
		Bool:   makeColoredProperty(color.FgHiMagenta),
		Number: makeColoredProperty(color.FgHiMagenta),
		MapKey: makeColoredProperty(color.FgHiCyan),
		Anchor: makeColoredProperty(color.FgHiYellow),
		Alias:  makeColoredProperty(color.FgHiYellow),
		String: makeColoredProperty(color.FgHiGreen),
	}

	tokens := lexer.Tokenize(marshaled)

	_, err := enc.w.Write([]byte(p.PrintTokens(tokens)))
	return err
}

// makeColoredProperty is a helper to produce color config for YAML property.
func makeColoredProperty(c color.Attribute) func() *printer.Property {
	return func() *printer.Property {
		return &printer.Property{
			Prefix: fmt.Sprintf(CSI+"%d"+CCT, c),
			Suffix: fmt.Sprintf(CSI+"%d"+CCT, color.Reset),
		}
	}
}

// NewColoredYAMLEncoder creates new colorful YAML encoder for given writer.
func NewColoredYAMLEncoder(w io.Writer, opts ...yaml.EncodeOption) *coloredYAMLEncoder {
	buffer := bytes.NewBuffer(nil)
	return &coloredYAMLEncoder{
		b:       buffer,
		Encoder: yaml.NewEncoder(buffer, opts...),
		w:       w,
	}
}

// UncolorizeReader is reserved for future use to uncolorize content of a reader.
func UncolorizeReader(r io.Reader) (io.Reader, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(colorSequenceRegex.ReplaceAll(raw, nil)), nil
}
