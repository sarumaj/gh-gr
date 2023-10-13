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
)

const pidFile = "gr.pid"

const ProcessAlreadyRunning = "gr is already running (pid: %d). Either kill the process or wait for it to terminate."

type processLockFile struct{ *os.File }

func (p processLockFile) Unlock() {
	pidFilePath := p.File.Name()
	util.FatalIfError(p.File.Close(), os.ErrClosed)
	util.FatalIfError(os.Remove(pidFilePath), os.ErrNotExist)
}

func AcquireProcessIDLock() interface{ Unlock() } {
	configDir := config.ConfigDir()
	pidFilePath := filepath.Join(configDir, pidFile)

	if util.PathExists(pidFilePath) {
		raw := util.FatalIfErrorOrReturn(os.ReadFile(pidFilePath))
		pid := util.FatalIfErrorOrReturn(strconv.Atoi(string(raw)))

		if proc, err := os.FindProcess(int(pid)); err == nil && !errors.Is(proc.Signal(syscall.Signal(0x0)), os.ErrProcessDone) {
			util.PrintlnAndExit(ProcessAlreadyRunning, proc.Pid)

		} else {
			util.FatalIfError(os.Remove(pidFilePath), os.ErrNotExist)

		}
	}

	f, err := os.OpenFile(pidFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	util.FatalIfError(err, os.ErrNotExist)

	_ = util.FatalIfErrorOrReturn(f.Write([]byte(fmt.Sprint(os.Getpid()))))

	return processLockFile{File: f}
}
