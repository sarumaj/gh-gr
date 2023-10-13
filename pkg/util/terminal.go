package util

import (
	"fmt"
	"os"
	"sync"

	term "github.com/cli/go-gh/v2/pkg/term"
)

var console = sync.Pool{New: func() any { return newConsoleTerminal() }}

type consoleTerminal struct{ *term.Term }

func (c *consoleTerminal) CheckColors(fn func(string, ...any) string, format string, a ...any) string {
	if c.ColorsEnabled() {
		return fn(format, a...)
	}

	return fmt.Sprintf(format, a...)
}

func (c *consoleTerminal) ColorsEnabled() bool {
	return c.IsTerminalOutput() && c.IsColorEnabled()
}

func (c *consoleTerminal) IsTerminal(out, err, in bool) bool {
	if out && !term.IsTerminal(c.Stdout()) {
		return false
	}

	if err && !term.IsTerminal(c.Stderr()) {
		return false
	}

	if in && !term.IsTerminal(c.Stdin()) { // will report false if stdin is a pipe
		return false
	}

	return out || err || in
}

func (c *consoleTerminal) Stderr() *os.File {
	return assertIsFileDescriptor(c.ErrOut(), os.Stderr)
}

func (c *consoleTerminal) Stdin() *os.File {
	return assertIsFileDescriptor(c.In(), os.Stdin)
}

func (c *consoleTerminal) Stdout() *os.File {
	return assertIsFileDescriptor(c.Out(), os.Stdout)
}

func assertIsFileDescriptor(w any, fallback *os.File) *os.File {
	v, ok := w.(*os.File)
	if !ok {
		return fallback
	}

	if v == nil {
		return fallback
	}

	return v
}

func Console() *consoleTerminal { return console.Get().(*consoleTerminal) }

func newConsoleTerminal() *consoleTerminal {
	t := term.FromEnv()
	return &consoleTerminal{&t}
}
