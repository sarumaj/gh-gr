package util

import (
	"bytes"
	"slices"
	"sync"

	tableprinter "github.com/cli/go-gh/v2/pkg/tableprinter"
	color "github.com/fatih/color"
	supererrors "github.com/sarumaj/go-super/errors"
)

type TablePrinter struct {
	sync.RWMutex
	isStdErr bool
	stdOut   tableprinter.TablePrinter
	stdErr   tableprinter.TablePrinter
	records  [][]string
}

func (t *TablePrinter) align(to int) {
	if to == 0 {
		to = len(t.records)
	}

	var maxLength int
	for rowId := 0; rowId < to; rowId++ {
		if len(t.records[rowId]) > maxLength {
			maxLength = len(t.records[rowId])
			if rowId > 0 {
				t.align(rowId)
			}

			continue
		}

		for len(t.records[rowId]) < maxLength {
			t.records[rowId] = append(t.records[rowId], "")
		}
	}
}

func (t *TablePrinter) current() tableprinter.TablePrinter {
	t.RLock()
	defer t.RUnlock()

	if t.isStdErr {
		return t.stdErr
	}

	return t.stdOut
}

func (t *TablePrinter) AddField(field string, colors ...color.Attribute) *TablePrinter {
	t.Lock()
	defer t.Unlock()

	if len(t.records) == 0 {
		t.records = append(t.records, nil)
	}

	c := Console()
	if len(colors) > 0 && c.ColorsEnabled() {
		t.records[len(t.records)-1] = append(t.records[len(t.records)-1], color.New(colors...).Sprint(field))
	} else {
		t.records[len(t.records)-1] = append(t.records[len(t.records)-1], field)
	}

	return t
}

func (t *TablePrinter) EndRow() *TablePrinter {
	t.Lock()
	defer t.Unlock()

	t.records = append(t.records, nil)

	return t
}

func (t *TablePrinter) Print() {
	current := t.current()

	t.align(0)
	for _, row := range t.records {
		for _, field := range row {
			current.AddField(field, tableprinter.WithTruncate(nil))
		}

		current.EndRow()
	}

	supererrors.Except(current.Render())
}

func (t *TablePrinter) Sprint() string {
	c := Console()
	width, _, _ := c.Size()
	width = max(width, 40)

	buffer := bytes.NewBuffer(nil)
	printer := tableprinter.New(buffer, c.IsTerminalOutput(), width)

	t.align(0)
	for _, row := range t.records {
		for _, field := range row {
			printer.AddField(field, tableprinter.WithTruncate(nil))
		}

		printer.EndRow()
	}

	supererrors.Except(printer.Render())

	return buffer.String()
}

func (t *TablePrinter) Sort() *TablePrinter {
	t.Lock()
	defer t.Unlock()

	slices.SortFunc(t.records, func(a, b []string) int {
		switch {
		case len(a)*len(b) > 0 && a[0] == b[0]:
			return 0

		case len(a)*len(b) > 0 && a[0] > b[0]:
			return 1

		case len(a)*len(b) > 0 && a[0] < b[0]:
			return -1

		default:
			return 0
		}
	})

	return t
}

func (t *TablePrinter) SetOutputToStdErr(isStdErr bool) *TablePrinter {
	t.Lock()
	defer t.Unlock()

	t.isStdErr = isStdErr
	return t
}

func NewTablePrinter() *TablePrinter {
	c := Console()
	width, _, _ := c.Size()
	width = max(width, 40)

	isTTY := c.IsTerminalOutput()
	return &TablePrinter{
		stdOut:  tableprinter.New(c.Stdout(), isTTY, width),
		stdErr:  tableprinter.New(c.Stderr(), isTTY, width),
		records: make([][]string, 1),
	}
}
