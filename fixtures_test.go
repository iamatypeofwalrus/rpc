package rpc

import (
	"context"
	"errors"
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

func (h *TestHandler) Handle(ctx context.Context, r *RequestTest) (*ResponseTest, error) {
	return &ResponseTest{Output: "success"}, nil
}

type TestHandlerFailure struct {
	err error
}

func (h *TestHandlerFailure) Input() *RequestTest {
	return &RequestTest{}
}

func (h *TestHandlerFailure) Output() *ResponseTest {
	return &ResponseTest{}
}

func (h *TestHandlerFailure) Handle(ctx context.Context, r *RequestTest) (*ResponseTest, error) {
	return nil, h.err
}

type badJSON struct{}

func (b badJSON) MarshalJSON() ([]byte, error) {
	return nil, errors.New("bad JSON")
}

type TestHandlerBadJSON struct{}

func (h TestHandlerBadJSON) Input() interface{} {
	return nil
}

func (h TestHandlerBadJSON) Output() interface{} {
	return badJSON{}
}

func (h TestHandlerBadJSON) Handle(ctx context.Context, input interface{}) (interface{}, error) {
	return h.Output(), nil
}
