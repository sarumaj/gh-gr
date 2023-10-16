package commands

import (
	"fmt"

	color "github.com/fatih/color"
	util "github.com/sarumaj/gh-gr/pkg/util"
)

// Utilizes table printer.
type operationStatus struct{ *util.TablePrinter }

// Append custom row.
func (p operationStatus) appendCustomRow(name string, args ...any) {
	_ = p.AddField(name)

	for _, status := range args {
		switch v := status.(type) {

		case string:
			p.AddField(v, color.FgGreen)

		case error:
			p.AddField(v.Error(), color.FgRed)

		default:
			p.AddField(fmt.Sprint(v))

		}
	}

	_ = p.EndRow()
}

// Append error row.
func (p operationStatus) appendErrorRow(name string, err error) {
	if err == nil {
		return
	}

	_ = p.AddField(name).AddField(err.Error(), color.FgRed).EndRow()
}

// Append status row.
func (p operationStatus) appendStatusRow(name, status string) {
	_ = p.AddField(name).AddField(status, color.FgGreen).EndRow()
}

// Initialize operation status.
func newOperationStatus() *operationStatus {
	return &operationStatus{util.NewTablePrinter()}
}
