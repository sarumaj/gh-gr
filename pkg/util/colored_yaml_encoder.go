package util

import (
	"bytes"
	"fmt"
	"io"
	"regexp"

	"github.com/fatih/color"
	yaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/printer"
)

// Regular expression for matching ANSI code sequences for color codes.
var colorSequenceRegex = regexp.MustCompile(regexp.QuoteMeta(CSI) + fmt.Sprintf("[^%[1]s]+%[1]s", CCT))

// Custom encoder to produce colored YAML output.
type coloredYAMLEncoder struct {
	b *bytes.Buffer
	w io.Writer
	*yaml.Encoder
}

// Encode YAML and colorize output.
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

// Helper to produce color config for YAML property.
func makeColoredProperty(c color.Attribute) func() *printer.Property {
	return func() *printer.Property {
		return &printer.Property{
			Prefix: fmt.Sprintf(CSI+"%d"+CCT, c),
			Suffix: fmt.Sprintf(CSI+"%d"+CCT, color.Reset),
		}
	}
}

// Create new colorful YAML encoder for given writer.
func NewColoredYAMLEncoder(w io.Writer, opts ...yaml.EncodeOption) *coloredYAMLEncoder {
	buffer := bytes.NewBuffer(nil)
	return &coloredYAMLEncoder{
		b:       buffer,
		Encoder: yaml.NewEncoder(buffer, opts...),
		w:       w,
	}
}

// For future use to uncolorize content of a reader.
func UncolorizeReader(r io.Reader) (io.Reader, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(colorSequenceRegex.ReplaceAll(raw, nil)), nil
}
