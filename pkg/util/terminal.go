package util

import (
	"fmt"
	"os"

	term "github.com/cli/go-gh/v2/pkg/term"
)

func Console() term.Term { return term.FromEnv() }

func CheckColors(fn func(string, ...any) string, format string, a ...any) string {
	if ColorsEnabled() {
		return fn(format, a...)
	}

	return fmt.Sprintf(format, a...)
}

func ColorsEnabled() bool {
	c := Console()
	return c.IsTerminalOutput() && c.IsColorEnabled()
}

func IsTerminal(out, err, in bool) bool {
	if out && !term.IsTerminal(Stdout()) {
		return false
	}

	if err && !term.IsTerminal(Stderr()) {
		return false
	}

	if in && !term.IsTerminal(Stdin()) { // will report false if unread content is present on stdin
		return false
	}

	return out || err || in
}

func Stderr() *os.File { return Console().ErrOut().(*os.File) }
func Stdin() *os.File  { return Console().In().(*os.File) }
func Stdout() *os.File { return Console().Out().(*os.File) }
