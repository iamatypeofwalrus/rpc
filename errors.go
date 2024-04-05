package rpc

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gorilla/schema"
)

var InternalServerError = RpcError{HTTPCode: 500, Message: "Internal Server Error"}

type RpcError struct {
	HTTPCode int
	Message  string
}

func (e RpcError) IsClientError() bool {
	return e.HTTPCode >= 400 && e.HTTPCode < 500
}

func (e RpcError) IsServerError() bool {
	return e.HTTPCode >= 500 && e.HTTPCode < 600
}

func (e RpcError) Error() string {
	return e.Message
}

func NewBadRequestError(message string) RpcError {
	return RpcError{HTTPCode: 400, Message: message}
}

func HandleError(w http.ResponseWriter, err error) {
	var httpErr RpcError
	if errors.As(err, &httpErr) {
		// assume that the error is an RpcError then at some point in the stack we've already logged it
		w.WriteHeader(httpErr.HTTPCode)
		json.NewEncoder(w).Encode(httpErr)
	} else {
		slog.Error("rpc error", "error", err)
		// If the error is not an RpcError, treat it as an internal server error
		w.WriteHeader(InternalServerError.HTTPCode)
		json.NewEncoder(w).Encode(InternalServerError)
	}
}

func handleUrlDecodeError(w http.ResponseWriter, err error) {
	emptyFieldError := &schema.MultiError{}
	if errors.As(err, emptyFieldError) {
		// If the error is an EmptyFieldError, return its message directly to the user
		HandleError(
			w,
			RpcError{
				HTTPCode: http.StatusBadRequest,
				Message:  emptyFieldError.Error(),
			},
		)
	} else {
		slog.Error("could not decode url params", "error", err)

		HandleError(
			w,
			RpcError{
				HTTPCode: http.StatusInternalServerError,
				Message:  "your request could not be processed",
			},
		)
	}
}
