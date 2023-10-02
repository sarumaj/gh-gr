package util

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/fatih/color"
)

func TestTeblePrinter(t *testing.T) {
	type text struct {
		text   string
		colors []color.Attribute
	}
	type args struct {
		input    []text
		isStdErr bool
	}
	for _, tt := range []struct {
		name string
		args args
		want string
	}{
		{"test#1", args{[]text{{"hello", nil}, {"world", []color.Attribute{}}}, false}, "hello\tworld\n"},
		{"test#2", args{[]text{{"hello", nil}, {"world", []color.Attribute{}}}, true}, "hello\tworld\n"},
	} {

		stderr, stdout := os.Stderr, os.Stdout

		stdoutname := filepath.Join(os.TempDir(), "stdout")
		stderrname := filepath.Join(os.TempDir(), "stderr")

		os.Stdout, _ = os.Create(stdoutname)
		os.Stderr, _ = os.Create(stderrname)

		defer func() {
			os.Stderr, os.Stdout = stderr, stdout
			_ = os.Remove(stdoutname)
			_ = os.Remove(stderrname)
		}()

		t.Run(tt.name, func(t *testing.T) {
			printer := TablePrinter().SetOutputToStdErr(tt.args.isStdErr)
			for _, arg := range tt.args.input {
				_ = printer.AddField(arg.text, arg.colors...)
			}

			if err := printer.EndRow().Render(); err != nil {
				t.Errorf(`TablePrinter()...Render() failed: %v`, err)
			}

			var got []byte
			if tt.args.isStdErr {
				got, _ = os.ReadFile(stderrname)
			} else {
				got, _ = os.ReadFile(stdoutname)
			}

			if !reflect.DeepEqual(got, []byte(tt.want)) {
				t.Errorf(`TablePrinter()...Render() failed: got: %q, want: %q`, got, tt.want)
			}
		})
	}
}
