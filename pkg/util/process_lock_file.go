package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	config "github.com/cli/go-gh/v2/pkg/config"
	supererrors "github.com/sarumaj/go-super/errors"
)

// Name of the PID file.
const pidFile = "gr.pid"

// Message when process with given ID is already running.
const ProcessAlreadyRunning = "gr is already running (pid: %d). Either kill the process or wait for it to terminate."

// Stores reference to the PID file.
type processLockFile struct{ *os.File }

// Release lock (close PID file and remove it).
func (p processLockFile) Unlock() {
	pidFilePath := p.File.Name()
	supererrors.Except(p.File.Close(), os.ErrClosed)
	supererrors.Except(os.Remove(pidFilePath), os.ErrNotExist)
}

// Record process ID and store it into the PID file.
// If process is already running, return corresponding message and exit.
func AcquireProcessIDLock() interface{ Unlock() } {
	configDir := config.ConfigDir()
	pidFilePath := filepath.Join(configDir, pidFile)

	if PathExists(pidFilePath) {
		raw := supererrors.ExceptFn(supererrors.W(os.ReadFile(pidFilePath)))
		pid := supererrors.ExceptFn(supererrors.W(strconv.Atoi(string(raw))))

		// correct way to check if process is running: send 0 signal
		if proc, err := os.FindProcess(int(pid)); err == nil && !errors.Is(proc.Signal(syscall.Signal(0x0)), os.ErrProcessDone) {
			PrintlnAndExit(ProcessAlreadyRunning, proc.Pid)

		} else {
			supererrors.Except(os.Remove(pidFilePath), os.ErrNotExist)

		}
	}

	f, err := os.OpenFile(pidFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	supererrors.Except(err, os.ErrNotExist)

	_ = supererrors.ExceptFn(supererrors.W(f.Write([]byte(fmt.Sprint(os.Getpid())))))

	return processLockFile{File: f}
}
