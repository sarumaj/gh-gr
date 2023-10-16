package commands

import "context"

// Custom implementation of context.
type operationContext struct {
	context.Context
}

// Used to store items in a context.
type operationContextKey string

// arguments to be stored in a context.
type operationContextMap map[operationContextKey]any

// Produce new operation context.
func newOperationContext(values operationContextMap) operationContext {
	ctx := &operationContext{context.Background()}

	for key, val := range values {
		ctx.Context = context.WithValue(ctx.Context, key, val)
	}

	return *ctx
}

// Retrieve element from context.
func unwrapOperationContext[T any](ctx operationContext, key operationContextKey) T {
	return ctx.Context.Value(key).(T)
}
