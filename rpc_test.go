package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetSuccess(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandler{}

	Register(mux, GET, "/test", handler)

	req, err := http.NewRequest(http.MethodGet, "/test?input=hello", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response ResponseTest
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}
	if response.Output != "success" {
		t.Errorf("Unexpected output value: got %s, want %s", response.Output, "success")
	}
}

func TestGetWithoutParams(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandler{}

	Register(mux, GET, "/test", handler)

	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestGetFailure(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandlerFailure{
		err: fmt.Errorf("error"),
	}

	Register(mux, GET, "/test", handler)

	req, err := http.NewRequest(http.MethodGet, "/test?input=hello", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

func TestGetBadJSON(t *testing.T) {
	mux := http.NewServeMux()
	handler := TestHandlerBadJSON{}

	Register(mux, GET, "/test", handler)

	req, err := http.NewRequest(http.MethodGet, "/test?input=hello", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

func TestPostSuccess(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandler{}

	Register(mux, POST, "/test", handler)

	reqBody := `{"input":"hello"}`
	req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response ResponseTest
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}
	if response.Output != "success" {
		t.Errorf("Unexpected output value: got %s, want %s", response.Output, "success")
	}

}

func TestPostWithoutBody(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandler{}

	Register(mux, POST, "/test", handler)

	req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(""))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestDefaultMiddleware(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandler{}

	// test that it was called and the order it was called
	d := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware")
			next.ServeHTTP(w, r)
		})
	}

	RegisterMiddleware(d)

	Register(mux, GET, "/test", handler)

	req, err := http.NewRequest(http.MethodGet, "/test?input=hello", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if rr.Header().Get("X-Test") != "middleware" {
		t.Errorf("Middleware was not applied")
	}

	// reset middleware
	defaultMiddleware = []Middleware{}
}

func TestRpcMiddleware(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandler{}

	// test that it was called and the order it was called
	r := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware")
			next.ServeHTTP(w, r)
		})
	}

	Register(mux, GET, "/test", handler, r)

	req, err := http.NewRequest(http.MethodGet, "/test?input=hello", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if rr.Header().Get("X-Test") != "middleware" {
		t.Errorf("Middleware was not applied")
	}

	// reset middleware
	defaultMiddleware = []Middleware{}
}

func TestDefaultAndRpcMiddlewareOrder(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandler{}
	header := "X-Test"

	num := 1
	first := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(
				header,
				fmt.Sprintf("%d", num),
			)
			next.ServeHTTP(w, r)
		})
	}

	num++
	second := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set(
				header,
				fmt.Sprintf("%d", num),
			)
			next.ServeHTTP(w, r)
		})
	}

	num++
	third := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set(
				header,
				fmt.Sprintf("%d", num),
			)
			next.ServeHTTP(w, r)
		})
	}

	RegisterMiddleware(first, second)

	Register(mux, GET, "/test", handler, third)

	req, err := http.NewRequest(http.MethodGet, "/test?input=hello", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	headerVal := rr.Header().Get("X-Test")
	if headerVal != "3" {
		t.Errorf("Middleware was not in the correct order, expected 3 got %s", headerVal)
	}

	// reset middleware
	defaultMiddleware = []Middleware{}
}

func TestHandlerNameMiddleware(t *testing.T) {
	mux := http.NewServeMux()
	middlewareTest := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerName := GetHandlerNameFromContext(r.Context())
			if handlerName != "TestHandler" {
				t.Errorf("Expected handler name to be 'TestHandler', got '%s'", handlerName)
			}
			next.ServeHTTP(w, r)
		})
	}

	Register(mux, GET, "/test", &TestHandler{}, middlewareTest)

	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
}

func TestDocsEndpoint(t *testing.T) {
	mux := http.NewServeMux()

	// Assuming you have a function to register your endpoints
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

	// Check the response body
	responseBody := rr.Body.String()
	if responseBody == "" {
		t.Errorf("Expected non-empty response body")
	}
}
