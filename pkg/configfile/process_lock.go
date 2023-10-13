package configfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	config "github.com/cli/go-gh/v2/pkg/config"
	util "github.com/sarumaj/gh-gr/pkg/util"
	supererrors "github.com/sarumaj/go-super/errors"
)

const pidFile = "gr.pid"

const ProcessAlreadyRunning = "gr is already running (pid: %d). Either kill the process or wait for it to terminate."

type processLockFile struct{ *os.File }

func (p processLockFile) Unlock() {
	pidFilePath := p.File.Name()
	supererrors.Except(p.File.Close(), os.ErrClosed)
	supererrors.Except(os.Remove(pidFilePath), os.ErrNotExist)
}

func AcquireProcessIDLock() interface{ Unlock() } {
	configDir := config.ConfigDir()
	pidFilePath := filepath.Join(configDir, pidFile)

	if util.PathExists(pidFilePath) {
		raw := supererrors.ExceptFn(supererrors.W(os.ReadFile(pidFilePath)))
		pid := supererrors.ExceptFn(supererrors.W(strconv.Atoi(string(raw))))

		if proc, err := os.FindProcess(int(pid)); err == nil && !errors.Is(proc.Signal(syscall.Signal(0x0)), os.ErrProcessDone) {
			util.PrintlnAndExit(ProcessAlreadyRunning, proc.Pid)

		} else {
			supererrors.Except(os.Remove(pidFilePath), os.ErrNotExist)

		}
	}

	f, err := os.OpenFile(pidFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	supererrors.Except(err, os.ErrNotExist)

	_ = supererrors.ExceptFn(supererrors.W(f.Write([]byte(fmt.Sprint(os.Getpid())))))

	return processLockFile{File: f}
}
