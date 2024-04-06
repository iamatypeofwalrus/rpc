package rpc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

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
	Handle(Input) (Output, error)
}

func Register[Input any, Output any](mux *http.ServeMux, httpMethod HTTPMethod, path string, h RpcHandler[Input, Output]) {
	muxPath := fmt.Sprintf("%s %s", httpMethod, path)

	// Register the handler to the mux
	mux.HandleFunc(muxPath, func(w http.ResponseWriter, r *http.Request) {
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

		resp, err := h.Handle(req)
		if err != nil {
			HandleError(w, err)
			return
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			HandleError(w, InternalServerError)
		}
	})

	registerHandlerForDocs(httpMethod, path, h)
}
