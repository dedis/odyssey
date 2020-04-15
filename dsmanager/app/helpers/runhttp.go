package helpers

import (
	"net/http"
)

// RunHTTP mocks traditional calls to an http server. This allows us to mock the
// enclave manager in the tests
type RunHTTP interface {
	Do(client *http.Client, req *http.Request) (*http.Response, error)
}

// DefaultHTTP is the default implementation of runhttp that uses net/http
type DefaultHTTP struct {
}

// Do implements runhttp by using the net/http lib
func (dh DefaultHTTP) Do(client *http.Client, req *http.Request) (resp *http.Response, err error) {
	return client.Do(req)
}

// NewDefaultRunHTTP returns the default RunHTTP
func NewDefaultRunHTTP() RunHTTP {
	return &DefaultHTTP{}
}
