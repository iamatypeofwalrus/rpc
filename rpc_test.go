package rpc

import (
	"encoding/json"
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
