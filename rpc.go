package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()
var defaultMiddleware []Middleware = []Middleware{}

func init() {
	decoder.IgnoreUnknownKeys(true)
}

type HTTPMethod string

const (
	GET  HTTPMethod = http.MethodGet
	POST HTTPMethod = http.MethodPost
)

type RpcHandler[Input any, Output any] interface {
	Input() Input
	Output() Output
	Handle(ctx context.Context, input Input) (Output, error)
}

// Register registers a new RPC handler to the given mux.
// handlers are applied in the order they are registered. Handlers regiestered with RegisterMiddleware are applied first.
func Register[Input any, Output any](mux *http.ServeMux, httpMethod HTTPMethod, path string, h RpcHandler[Input, Output], rpcMiddleware ...Middleware) {
	muxPath := fmt.Sprintf("%s %s", httpMethod, path)

	rpcHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
		req := h.Input()

		w.Header().Set("Content-Type", "application/json")

		// TODO support path parameters, too
		if httpMethod == GET || httpMethod == http.MethodDelete {
			query := r.URL.Query()

			if err := decoder.Decode(req, query); err != nil {
				handleUrlDecodeError(w, err)
				return
			}
		} else {
			defer r.Body.Close()

			if err := json.NewDecoder(r.Body).Decode(req); err != nil {
				slog.Error(
					"could not deserialize body",
					"error", err,
					"body", r.Body,
				)
				HandleError(
					w,
					RpcError{
						HTTPCode: http.StatusBadRequest,
						Message:  "Bad Request: invalid body",
					},
				)
				return
			}
		}

		resp, err := h.Handle(r.Context(), req)
		if err != nil {
			HandleError(w, err)
			return
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			HandleError(w, InternalServerError)
		}
	}

	rpcHandler := http.HandlerFunc(rpcHandlerFunc)

	clonedDefaultMiddleware := slices.Clone(defaultMiddleware)
	clonedDefaultMiddleware = append(clonedDefaultMiddleware, rpcMiddleware...)

	wrappedHandler := chainMiddleware(rpcHandler, clonedDefaultMiddleware)

	// Register the handler to the mux
	mux.Handle(muxPath, wrappedHandler)

	registerHandlerForDocs(httpMethod, path, h)
}

// RegisterMiddleware registers a middleware that is used for all RPCs that are registed.
// Use the handlers argument in Register to apply RPC specific middleware in addition to the default middleware.
//
// Handlers are applied in the order they are registered.
func RegisterMiddleware(middlewware ...Middleware) {
	defaultMiddleware = append(defaultMiddleware, middlewware...)
}

func chainMiddleware(final http.Handler, middleware []Middleware) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		final = middleware[i](final)
	}

	return final
}
