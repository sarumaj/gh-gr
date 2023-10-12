package configfile

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"time"

	config "github.com/cli/go-gh/v2/pkg/config"
	fslock "github.com/juju/fslock"
	util "github.com/sarumaj/gh-gr/pkg/util"
)

const pidFile = "gr.pid"

var pidFilePath = filepath.Join(config.ConfigDir(), pidFile)

type ProcessLock struct{ lock *fslock.Lock }

func (p ProcessLock) Lock(timeout time.Duration) {
	if timeout > 0 {
		util.FatalIfError(p.lock.LockWithTimeout(timeout))
		return
	}

	util.FatalIfError(p.lock.Lock())
}

func (p ProcessLock) Unlock() {
	util.FatalIfError(p.lock.Unlock())
	_ = os.Remove(pidFilePath)
}

func NewProcessLock() *ProcessLock {
	f, err := os.OpenFile(pidFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	util.FatalIfError(err)

	err = binary.Write(f, binary.LittleEndian, uint32(os.Getpid()))
	_ = f.Close()
	util.FatalIfError(err)

	return &ProcessLock{fslock.New(pidFilePath)}
}
