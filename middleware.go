package rpc

import (
	"context"
	"net/http"
	"reflect"
)

type Middleware func(http.Handler) http.Handler

type handlerNameContextKey struct{}

func handlerNameMiddleware[Input any, Output any](h RpcHandler[Input, Output]) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			handlerType := reflect.TypeOf(h)

			// If the RpcHandler is a pointer, get the type it points to
			if handlerType.Kind() == reflect.Ptr {
				handlerType = handlerType.Elem()
			}
			// Get the handler name
			handlerName := handlerType.Name()

			// Add the handler name to the context
			ctx := SetHandlerNameInContext(r.Context(), handlerName)

			// Call the next middleware or handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SetHandlerNameInContext(ctx context.Context, handlerName string) context.Context {
	return context.WithValue(ctx, handlerNameContextKey{}, handlerName)
}

func GetHandlerNameFromContext(ctx context.Context) string {
	if handlerName, ok := ctx.Value(handlerNameContextKey{}).(string); ok {
		return handlerName
	}
	return ""
}
