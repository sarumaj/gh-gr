package util

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime"
	"sync"

	color "github.com/fatih/color"
)

var interruptBlocker = sync.Pool{
	New: func() any {
		return &blocker{
			signal: make(chan os.Signal, 1),
			quit:   make(chan bool, 1),
		}
	},
}

type blocker struct {
	sync.Mutex
	signal chan os.Signal
	quit   chan bool
}

func (i *blocker) Fire() {
	for !i.TryLock() {
	}

	defer signal.Notify(i.signal, os.Interrupt)

	go func(interrupt <-chan os.Signal, quit <-chan bool) {
		for c := Console(); ; {
			select {

			case <-interrupt:
				_ = FatalIfErrorOrReturn(
					fmt.Fprintln(c.Stderr(), c.CheckColors(color.RedString, c.CheckColors(color.RedString, "Current execution cannot be interrupted!"))),
				)

			case <-quit:
				return

			}
		}
	}(i.signal, i.quit)
}

func (i *blocker) Stop() {
	defer i.Unlock()

	i.quit <- true
	signal.Stop(i.signal)
}

func GetIdealConcurrency() uint {
	return uint(math.Max(float64(runtime.NumCPU()*2), 4))
}

func PreventInterrupt() interface{ Stop() } {
	i := interruptBlocker.Get().(*blocker)
	defer i.Fire()
	return i
}
