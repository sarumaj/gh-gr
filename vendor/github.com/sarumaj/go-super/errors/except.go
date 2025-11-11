package errors

import "errors"

// Handle error if not nil, and not among ignored ones.
func Except(err error, ignore ...error) {
	if err == nil {
		return
	}

	lastError.store(err)

	for _, e := range ignore {
		if errors.Is(err, e) {
			return
		}
	}

	callback.fn(err)
}

// Handle error if not nil, and not among ignored ones.
// Return anything from fn except for error if successful.
func ExceptFn[T any](fn ErrorFn[T], ignore ...error) T {
	t, err := fn()
	Except(err, ignore...)
	return t
}

// Handle error if not nil, and not among ignored ones.
// Return anything from fn except for error if successful.
func ExceptFn2[T, U any](fn ErrorFn2[T, U], ignore ...error) (T, U) {
	t, u, err := fn()
	Except(err, ignore...)
	return t, u
}
