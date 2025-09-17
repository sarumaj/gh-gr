package util

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime"
	"sync"

	color "github.com/fatih/color"
	supererrors "github.com/sarumaj/go-super/errors"
)

// Instance of interrupt blocker (thread-safe storage).
var interruptBlocker = sync.Pool{
	New: func() any {
		return &blocker{
			signal: make(chan os.Signal, 1),
			quit:   make(chan bool, 1),
		}
	},
}

// Blocker to catch interrupt signals and handle them,
type blocker struct {
	sync.Mutex
	signal chan os.Signal
	quit   chan bool
}

// Dispatch watcher to listen for incoming SIGINT signals.
func (i *blocker) Fire() {
	for !i.TryLock() {
	}

	defer signal.Notify(i.signal, os.Interrupt)

	go func(interrupt <-chan os.Signal, quit <-chan bool) {
		for c := Console(); ; {
			select {

			case <-interrupt:
				_ = supererrors.ExceptFn(supererrors.W(
					fmt.Fprintln(c.Stderr(), c.CheckColors(color.RedString, "Current execution cannot be interrupted!")),
				))

			case <-quit:
				return

			}
		}
	}(i.signal, i.quit)
}

// Stop watcher.
func (i *blocker) Stop() {
	defer i.Unlock()

	i.quit <- true
	signal.Stop(i.signal)
}

// Get ideal number of workers for current CPU.
func GetIdealConcurrency() uint {
	return uint(math.Max(float64(runtime.NumCPU()*2), 4))
}

// Dispatch watcher.
func PreventInterrupt() interface{ Stop() } {
	i := interruptBlocker.Get().(*blocker)
	defer i.Fire()
	return i
}
