package errors

type (
	// Function of type func[T any]() (T, error).
	ErrorFn[T any] func() (T, error)

	// Function of type func[T, U any]() (T, U, error).
	ErrorFn2[T, U any] func() (T, U, error)
)

// Wrapper for function of type func[T any]() (T, error).
func W[T any](t T, err error) ErrorFn[T] { return func() (T, error) { return t, err } }

// Wrapper for function of type func[T, U any]() (T, U, error).
func W2[T, U any](t T, u U, err error) ErrorFn2[T, U] {
	return func() (T, U, error) { return t, u, err }
}
