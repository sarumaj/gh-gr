package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/cli/safeexec"
)

func TestExec(t *testing.T) {
	goBinary, err := safeexec.LookPath("go")
	if err != nil {
		t.Skipf("go executable not found, skipping: %v", err)
	}

	version := "v0.0.0"
	buildDate := time.Time{}.Format(`2006-01-02 15:04:05 MST`)

	target := filepath.Join(os.TempDir(), "gh-gr")
	if runtime.GOOS == "windows" {
		target += ".exe"
	}

	defer os.Remove(target)

	cmd := exec.Command(
		goBinary,
		"build",
		"-trimpath",
		"-ldflags=-s -w -X 'main.Version="+version+"' -X 'main.BuildDate="+buildDate+"' -extldflags=-static",
		"-tags=osusergo netgo static_build",
		"-o",
		target,
		"main.go",
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("exec.Command(%q, \"build\", ..., \"-o\", %q, \"main.go\") failed: %v, output:\n%s", goBinary, target, err, out)
	}

	targetBinary, err := safeexec.LookPath(target)
	if err != nil {
		t.Errorf(`safeexec.LookPath(%q) failed: %v`, target, err)
	}

	cmd = exec.Command(targetBinary, "version")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("exec.Command(%q, \"version\") failed: %v, output:\n%s", targetBinary, err, out)
	}
}
