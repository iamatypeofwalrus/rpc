package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type RequestTest struct {
	Input string `schema:"input,required" json:"input"`
}

type ResponseTest struct {
	Output string `json:"output"`
}

type TestHandler struct{}

func (h *TestHandler) Input() *RequestTest {
	return &RequestTest{}
}

func (h *TestHandler) Output() *ResponseTest {
	return &ResponseTest{}
}

func (h *TestHandler) Handle(RequestTest *RequestTest) (*ResponseTest, error) {
	return &ResponseTest{Output: "success"}, nil
}

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
