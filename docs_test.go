package rpc

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterDocsSetsContext(t *testing.T) {
	mux := http.NewServeMux()

	middleware := func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value := GetHandlerNameFromContext(r.Context())
			if value == "" {
				t.Error("Context does not contain RpcDocs")
			}
			handler.ServeHTTP(w, r)
		})
	}

	RegisterMiddleware(middleware)
	defer resetDefaultMiddleware()

	RegisterDocsEndpoint(mux)

	req, err := http.NewRequest(http.MethodGet, "/rpc/docs", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
