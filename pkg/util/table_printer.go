package util

import (
	"bytes"
	"slices"

	tableprinter "github.com/cli/go-gh/v2/pkg/tableprinter"
	color "github.com/fatih/color"
	supererrors "github.com/sarumaj/go-super/errors"
)

// Table printer prints text in padded, tabular form.
type TablePrinter struct {
	isStdErr bool
	stdOut   tableprinter.TablePrinter
	stdErr   tableprinter.TablePrinter
	records  [][]string
}

// Align records to match the maximum column width and prevent index errors.
func (t *TablePrinter) Align() *TablePrinter {
	var max_row_width int
	for _, row := range t.records {
		if len(row) > max_row_width {
			max_row_width = len(row)
		}
	}

	for i := range t.records {
		for len(t.records[i]) < max_row_width {
			t.records[i] = append(t.records[i], "")
		}
	}

	return t
}

// Get current printer.
func (t *TablePrinter) current() tableprinter.TablePrinter {
	if t.isStdErr {
		return t.stdErr
	}

	return t.stdOut
}

// Add record.
func (t *TablePrinter) AddRowField(field string, colors ...color.Attribute) *TablePrinter {
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

// End row (new line).
func (t *TablePrinter) EndRow() *TablePrinter {
	t.records = append(t.records, nil)
	return t
}

// Print to Stdout.
func (t *TablePrinter) Print() {
	current := t.current()

	for _, row := range t.records {
		for _, field := range row {
			current.AddField(field)
		}

		current.EndRow()
	}

	supererrors.Except(current.Render())
}

// Print through a buffer to deliver rendered string.
func (t *TablePrinter) Sprint() string {
	c := Console()
	width, _, _ := c.Size()
	width = max(width, 40)

	buffer := bytes.NewBuffer(nil)
	printer := tableprinter.New(buffer, c.IsTerminalOutput(), width)

	for _, row := range t.records {
		for _, field := range row {
			printer.AddField(field)
		}

		printer.EndRow()
	}

	supererrors.Except(printer.Render())

	return buffer.String()
}

// Sort records.
func (t *TablePrinter) Sort() *TablePrinter {
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

// Switch between Stdout and Stderr.
func (t *TablePrinter) SetOutputToStdErr(isStdErr bool) *TablePrinter {
	t.isStdErr = isStdErr
	return t
}

// Create new table printer.
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
