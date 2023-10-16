package util

import (
	"fmt"
	"os"
	"sync"

	term "github.com/cli/go-gh/v2/pkg/term"
)

// Stores terminal information unit (thread-safe).
var console = sync.Pool{New: func() any { return newConsoleTerminal() }}

// Custom type for terminal information.
type consoleTerminal struct{ *term.Term }

// If ANSI color codes are supported will print colorful text, otherwise without color.
func (c *consoleTerminal) CheckColors(fn func(string, ...any) string, format string, a ...any) string {
	if c.ColorsEnabled() {
		return fn(format, a...)
	}

	return fmt.Sprintf(format, a...)
}

// Check if terminal supports ANSI color codes.
func (c *consoleTerminal) ColorsEnabled() bool {
	return c.IsTerminalOutput() && c.IsColorEnabled()
}

// Check if terminal is writeable (Stdout, Stderr and Stdin).
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

// Get file descriptor for Stderr.
func (c *consoleTerminal) Stderr() *os.File {
	return assertIsFileDescriptor(c.ErrOut(), os.Stderr)
}

// Get file descriptor for Stdin.
func (c *consoleTerminal) Stdin() *os.File {
	return assertIsFileDescriptor(c.In(), os.Stdin)
}

// Get file descriptor for Stdout.
func (c *consoleTerminal) Stdout() *os.File {
	return assertIsFileDescriptor(c.Out(), os.Stdout)
}

// Make sure retrieved file descriptor for terminal is valid.
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

// Access point for terminal information.
func Console() *consoleTerminal { return console.Get().(*consoleTerminal) }

// Create new terminal information unit.
func newConsoleTerminal() *consoleTerminal {
	t := term.FromEnv()
	return &consoleTerminal{&t}
}
