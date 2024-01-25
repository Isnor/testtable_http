package testtablehttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// JSONEndpointTest is a kind of testtable.Test whose run function executes an endpoint and runs
// assertions against the response
type JSONEndpointTest[RequestType, ResponseType any] struct {
	Name         string
	Path         string
	HTTPMethod   string
	RequestBody  *RequestType
	Handler      http.HandlerFunc
	Expectations func(parent *testing.T, httpResponse *http.Response, response *ResponseType)
}

// Run runs this `test` by marshaling its `Input` and using an `httptest.ResponseRecorder`
// to execute its `Handler`.
func (test *JSONEndpointTest[RequestType, ResponseType]) Run(t *testing.T) {
	// turn our request/test input into an HTTP request
	request := httptest.NewRequest(test.HTTPMethod, test.Path, nil)
	if test.RequestBody != nil {
		body, err := json.Marshal(test.RequestBody)
		if err != nil {
			t.Errorf("error writing post request body %v %v\n", err, test.RequestBody)
			return
		}
		request = httptest.NewRequest(test.HTTPMethod, test.Path, bytes.NewBuffer(body))
	}

	recorder := httptest.NewRecorder()
	test.Handler(recorder, request)
	endpointResult := new(ResponseType)
	httpResponse := recorder.Result()
	// unmarshal the response body - we're assuming that the (converted) handler has encoded a JSON body
	if err := json.NewDecoder(httpResponse.Body).Decode(endpointResult); err != nil {
		t.Errorf("failed deserializing response body %v\n", err)
	}

	// run assertions on the output
	test.Expectations(t, httpResponse, endpointResult)
}

// below are convenience functions that try to improve the quality of life of defining tests by infering
// the generic type based on the arguments of the function. Go supports this with functions, but as of
// 1.18 it does not with struct types

// GetTest returns a JSONEndpointTest that has no Input / body and uses http.MethodGet for the HTTPMethod
func GetTest[ResponseType any](
	name string,
	path string,
	handler http.HandlerFunc,
	expectations func(parent *testing.T, httpResponse *http.Response, response *ResponseType),
) *JSONEndpointTest[any, ResponseType] {
	return &JSONEndpointTest[any, ResponseType]{
		Name:         name,
		Path:         path,
		HTTPMethod:   http.MethodGet,
		RequestBody:  nil,
		Handler:      handler,
		Expectations: expectations,
	}
}

// PostTest returns a JSONEndpointTest that uses http.MethodPOST for the HTTPMethod
func PostTest[RequestType, ResponseType any](
	name string,
	path string,
	input *RequestType,
	handler http.HandlerFunc,
	expectations func(parent *testing.T, httpResponse *http.Response, response *ResponseType),
) *JSONEndpointTest[RequestType, ResponseType] {
	return &JSONEndpointTest[RequestType, ResponseType]{
		Name:         name,
		Path:         path,
		HTTPMethod:   http.MethodPost,
		RequestBody:  input,
		Handler:      handler,
		Expectations: expectations,
	}
}
