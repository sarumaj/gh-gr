package commands

import (
	"fmt"
	"sync"

	color "github.com/fatih/color"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
)

// Utilizes table printer.
type operationStatus struct {
	sync.Mutex
	*util.TablePrinter
}

// Append custom row.
func (p *operationStatus) appendRow(name string, args ...any) {
	p.Lock()
	defer p.Unlock()

	_ = p.AddRowField(name)

	for _, status := range args {
		if status == nil {
			continue
		}

		switch v := status.(type) {

		case string:
			p.AddRowField(v, color.FgGreen)
		case error:
			p.AddRowField(v.Error(), color.FgRed)

		default:
			p.AddRowField(fmt.Sprint(v))

		}
	}

	_ = p.EndRow()
}

// Initialize operation status.
func newOperationStatus() *operationStatus {
	return &operationStatus{TablePrinter: util.NewTablePrinter()}
}
