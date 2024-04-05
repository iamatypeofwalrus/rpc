package rpc

import (
	"encoding/json"
	"net/http"
	"reflect"
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
	mux.HandleFunc("/rpc/docs", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(rpcRegistry)
	})
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
