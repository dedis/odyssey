package helpers

import (
	"net/http"
	"net/url"
)

// RunHTTP mocks traditional calls to an http server. This allows us to mock the
// enclave manager in the tests
type RunHTTP interface {
	PostForm(url string, data url.Values) (resp *http.Response, err error)
}

// DefaultHTTP is the default implementation of runhttp that uses net/http
type DefaultHTTP struct {
}

// PostForm implements runhttp by using the net/http lib
func (dh DefaultHTTP) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return http.PostForm(url, data)
}

// NewDefaultRunHTTP returns the default RunHTTP
func NewDefaultRunHTTP() RunHTTP {
	return &DefaultHTTP{}
}
