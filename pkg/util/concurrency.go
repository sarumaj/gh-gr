package util

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/fatih/color"
)

type interrupt struct {
	signal chan os.Signal
	quit   chan bool
}

func (i *interrupt) Stop() {
	i.quit <- true
	signal.Stop(i.signal)
}

func GetIdealConcurrency() uint {
	return uint(math.Max(float64(runtime.NumCPU()*2), 4))
}

func NewInterrupt() *interrupt {
	i := &interrupt{
		signal: make(chan os.Signal, 1),
		quit:   make(chan bool, 1),
	}

	signal.Notify(i.signal, os.Interrupt, syscall.SIGTERM)

	go func(interrupt <-chan os.Signal, quit <-chan bool) {
		for {
			select {

			case <-interrupt:
				fmt.Fprintln(os.Stderr, CheckColors(color.RedString, CheckColors(color.RedString, "\nCurrent execution cannot be interrupted!")))

			case <-quit:
				return
			}
		}
	}(i.signal, i.quit)

	return i
}
