package errors

import (
	"fmt"
	"os"
)

// Storage for callback function.
var callback = (&defaultCallback{}).reset()

// Store callback function
type defaultCallback struct {
	fn Callback
}

// Reset callback function to fmt.Fprintln(os.Stderr, err).
// Used for initialization as well.
func (fn *defaultCallback) reset() *defaultCallback {
	fn.fn = func(err error) {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}

	return fn
}

// Callback function to handle error.
type Callback func(error)

// Reset callback function to fmt.Fprintln(os.Stderr, err).
func RestoreCallback() {
	callback.reset()
}

// Register custom callback to handle error.
func RegisterCallback(fn Callback) {
	callback.fn = fn
}
