package util

import (
	tableprinter "github.com/cli/go-gh/v2/pkg/tableprinter"
	color "github.com/fatih/color"
)

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

	if len(colors) > 0 && ColorsEnabled() {
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

func (t *tablePrinter) Render() {
	current := t.current()
	FatalIfError(current.Render())
}

func (t *tablePrinter) SetOutputToStdErr(isStdErr bool) *tablePrinter {
	t.isStdErr = isStdErr
	return t
}

func NewTablePrinter() *tablePrinter {
	c := Console()
	width, _, _ := c.Size()
	width = max(width, 40)

	isTTy := c.IsTerminalOutput()
	return &tablePrinter{
		stdOut: tableprinter.New(Stdout(), isTTy, width),
		stdErr: tableprinter.New(Stderr(), isTTy, width),
	}
}
