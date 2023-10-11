package commands

import "context"

type operationContext struct {
	context.Context
}

type operationContextKey string

type operationContextMap map[operationContextKey]any

func newOperationContext(values operationContextMap) operationContext {
	ctx := &operationContext{context.Background()}

	for key, val := range values {
		ctx.Context = context.WithValue(ctx.Context, key, val)
	}

	return *ctx
}

func unwrapOperationContext[T any](ctx operationContext, key operationContextKey) T {
	return ctx.Context.Value(key).(T)
}
