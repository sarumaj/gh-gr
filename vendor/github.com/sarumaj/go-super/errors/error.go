package errors

import (
	"errors"
	"sync"
)

// Storer for last occurred error.
var lastError = &errorKeeper{}

// Store last error (thread-safe).
type errorKeeper struct {
	err error
	sync.RWMutex
}

// Retrieve last error
func (k *errorKeeper) read() error {
	for !k.TryRLock() {
	}
	defer k.RUnlock()

	return k.err
}

// Store error.
func (k *errorKeeper) store(err error) {
	for !k.TryLock() {
	}
	k.err = err
	k.Unlock()
}

// Retrieve last error.
func LastError() error {
	return lastError.read()
}

// Check if last error was of this kind.
func LastErrorWas(err error) bool {
	return errors.Is(lastError.read(), err)
}
