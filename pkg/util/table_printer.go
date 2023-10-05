package util

import (
	"fmt"
	"os"
	"sync"

	tableprinter "github.com/cli/go-gh/v2/pkg/tableprinter"
	term "github.com/cli/go-gh/v2/pkg/term"
	color "github.com/fatih/color"
)

var printer = sync.Pool{New: newTablePrinter}

type tablePrinter struct {
	isStdErr bool
	stdOut   tableprinter.TablePrinter
	stdErr   tableprinter.TablePrinter
}

func (t *tablePrinter) current() tableprinter.TablePrinter {
	if t.isStdErr {
		return t.stdErr
	}

	return t.stdOut
}

func (t *tablePrinter) AddField(field string, colors ...color.Attribute) *tablePrinter {
	current := t.current()

	if len(colors) > 0 && UseColors() {
		current.AddField(
			field,
			tableprinter.WithColor(func(s string) string {
				return color.New(colors...).Sprint(s)
			}),
			tableprinter.WithTruncate(nil),
		)

		return t
	}

	current.AddField(field, tableprinter.WithTruncate(nil))
	return t
}

func (t *tablePrinter) EndRow() *tablePrinter {
	current := t.current()
	current.EndRow()
	return t
}

func (t *tablePrinter) Render() error {
	current := t.current()
	return current.Render()
}

func (t *tablePrinter) SetOutputToStdErr(isStdErr bool) *tablePrinter {
	t.isStdErr = isStdErr
	return t
}

func CheckColors(fn func(string, ...any) string, format string, a ...any) string {
	if UseColors() {
		return fn(format, a...)
	}

	return fmt.Sprintf(format, a...)
}

func newTablePrinter() any {
	console := term.FromEnv()

	width, _, _ := console.Size()
	width = max(width, 40)

	isTTy := console.IsTerminalOutput()
	return &tablePrinter{
		stdOut: tableprinter.New(os.Stdout, isTTy, width),
		stdErr: tableprinter.New(os.Stderr, isTTy, width),
	}
}

func UseColors() bool {
	console := term.FromEnv()

	isTTY := console.IsTerminalOutput()
	colorsOn := console.IsColorEnabled()

	return isTTY && colorsOn
}

func TablePrinter() *tablePrinter { return printer.Get().(*tablePrinter) }
