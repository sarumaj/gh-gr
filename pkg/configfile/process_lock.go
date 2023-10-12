package configfile

import (
	"os"
	"path/filepath"

	config "github.com/cli/go-gh/v2/pkg/config"
	"github.com/go-git/go-git/v5/utils/binary"
	util "github.com/sarumaj/gh-gr/pkg/util"
)

const pidFile = "gr.pid"

const ProcessAlreadyRunning = "gr is already running (pid: %d). Either kill the process or wait for it to terminate."

type ProcessLockFile struct{ *os.File }

func (p ProcessLockFile) Unlock() {
	path := p.File.Name()
	util.FatalIfError(p.File.Close())
	util.FatalIfError(os.Remove(path))
}

func AquireProcessIDLock(kill bool) ProcessLockFile {
	configDir := config.ConfigDir()
	pidFilePath := filepath.Join(configDir, pidFile)

	if util.PathExists(pidFilePath) {
		f, err := os.Open(pidFilePath)
		util.FatalIfError(err)

		pid, err := binary.ReadVariableWidthInt(f)
		util.FatalIfError(err)

		_ = f.Close()
		if kill {
			util.FatalIfError(os.Remove(pidFilePath))
			if proc, err := os.FindProcess(int(pid)); err == nil {
				_ = proc.Kill()
			}
		} else {
			util.PrintlnAndExit(ProcessAlreadyRunning, pid)
		}
	}

	f, err := os.OpenFile(filepath.Join(configDir, pidFile), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	util.FatalIfError(err)
	util.FatalIfError(binary.WriteVariableWidthInt(f, int64(os.Getpid())))

	return ProcessLockFile{f}
}
