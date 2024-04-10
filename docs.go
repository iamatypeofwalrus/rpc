package rpc

import (
	"encoding/json"
	"net/http"
	"reflect"
	"slices"
)

var goToJsonTypes = map[string]string{
	"string":  "string",
	"int":     "number",
	"int32":   "number",
	"int64":   "number",
	"float32": "number",
	"float64": "number",
	"bool":    "boolean",
}

// Docs
type RpcEndpoint struct {
	Path       string      `json:"path"`
	HTTPMethod HTTPMethod  `json:"httpMethod"`
	InputType  string      `json:"inputType"`
	Input      interface{} `json:"input"`
	OutputType string      `json:"outputType"`
	Output     interface{} `json:"output"`
}

var rpcRegistry []RpcEndpoint

func RegisterDocsEndpoint(mux *http.ServeMux) {
	contextMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			SetHandlerNameInContext(r.Context(), "RpcDocs")
			next.ServeHTTP(w, r)
		})
	}

	middleware := slices.Clone(defaultMiddleware)
	middleware = append([]Middleware{contextMiddleware}, middleware...)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(rpcRegistry)
	})

	mux.Handle(
		"/rpc/docs",
		chainMiddleware(h, middleware...),
	)
}

func registerHandlerForDocs[Input any, Output any](httpMethod HTTPMethod, path string, h RpcHandler[Input, Output]) {
	var inputType string
	if httpMethod == GET || httpMethod == http.MethodDelete {
		inputType = "query params"
	} else {
		inputType = "json body"
	}

	input := h.Input()
	output := h.Output()

	inputFields := getFields(input)
	outputFields := getFields(output)

	rpcRegistry = append(rpcRegistry, RpcEndpoint{
		Path:       path,
		HTTPMethod: httpMethod,
		InputType:  inputType,
		Input:      inputFields,
		OutputType: "json body",
		Output:     outputFields,
	})
}

func getFields(i interface{}) map[string]string {
	if i == nil {
		return nil
	}

	fields := make(map[string]string)
	v := reflect.ValueOf(i)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		}

		goType := field.Type.String()
		jsonType, ok := goToJsonTypes[goType]
		if !ok {
			jsonType = "object" // default to "object" if the Go type is not in the map
		}
		fields[jsonTag] = jsonType
	}

	return fields
}
