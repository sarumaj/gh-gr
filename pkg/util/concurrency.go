package util

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	color "github.com/fatih/color"
)

var interruptInstance = &interrupt{
	signal: make(chan os.Signal, 1),
	quit:   make(chan bool, 1),
}

type interrupt struct {
	sync.Mutex
	signal chan os.Signal
	quit   chan bool
}

func (i *interrupt) Fire() {
	for !i.TryLock() {
	}

	defer signal.Notify(i.signal, os.Interrupt, syscall.SIGTERM)

	go func(interrupt <-chan os.Signal, quit <-chan bool) {
		for {
			select {

			case <-interrupt:
				_, _ = fmt.Fprintln(Stderr(), CheckColors(color.RedString, CheckColors(color.RedString, "Current execution cannot be interrupted!")))

			case <-quit:
				return

			}
		}
	}(i.signal, i.quit)
}

func (i *interrupt) Stop() {
	defer i.Unlock()

	i.quit <- true
	signal.Stop(i.signal)
}

func GetIdealConcurrency() uint {
	return uint(math.Max(float64(runtime.NumCPU()*2), 4))
}

func PreventInterrupt() interface{ Stop() } {
	defer interruptInstance.Fire()
	return interruptInstance
}
