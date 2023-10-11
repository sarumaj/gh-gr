package util

import (
	"sort"

	tableprinter "github.com/cli/go-gh/v2/pkg/tableprinter"
	color "github.com/fatih/color"
)

type TablePrinter struct {
	isStdErr bool
	stdOut   tableprinter.TablePrinter
	stdErr   tableprinter.TablePrinter
	records  [][]string
}

func (t *TablePrinter) current() tableprinter.TablePrinter {
	if t.isStdErr {
		return t.stdErr
	}

	return t.stdOut
}

func (t *TablePrinter) AddField(field string, colors ...color.Attribute) *TablePrinter {
	if len(t.records) == 0 {
		t.records = append(t.records, nil)
	}

	if len(colors) > 0 && ColorsEnabled() {
		t.records[len(t.records)-1] = append(t.records[len(t.records)-1], color.New(colors...).Sprint(field))
	} else {
		t.records[len(t.records)-1] = append(t.records[len(t.records)-1], field)
	}

	return t
}

func (t *TablePrinter) EndRow() *TablePrinter {
	t.records = append(t.records, nil)

	return t
}

func (t *TablePrinter) Print() {
	current := t.current()

	for _, row := range t.records {
		for _, field := range row {
			current.AddField(field, tableprinter.WithTruncate(nil))
		}

		current.EndRow()
	}

	FatalIfError(current.Render())
}

func (t *TablePrinter) Sort() *TablePrinter {
	sort.Slice(t.records, func(i, j int) bool {
		return true &&
			len(t.records[i]) > 0 &&
			len(t.records[j]) > 0 &&
			t.records[i][0] < t.records[j][0]
	})

	return t
}

func (t *TablePrinter) SetOutputToStdErr(isStdErr bool) *TablePrinter {
	t.isStdErr = isStdErr
	return t
}

func NewTablePrinter() *TablePrinter {
	c := Console()
	width, _, _ := c.Size()
	width = max(width, 40)

	isTTy := c.IsTerminalOutput()
	return &TablePrinter{
		stdOut:  tableprinter.New(Stdout(), isTTy, width),
		stdErr:  tableprinter.New(Stderr(), isTTy, width),
		records: make([][]string, 1),
	}
}
