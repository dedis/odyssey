package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/sessions"
	"go.dedis.ch/onet/v3/log"
	"golang.org/x/xerrors"
)

/*
	This file holds the utilities for error handling of http requests
*/

// RequestError is returned as json formated when something went wrong during
// the handling of a request
type RequestError struct {
	Message string
	Code    int
	Status  string
	Data    map[string]interface{}
}

func (e RequestError) Error() string {
	return fmt.Sprintf("request error %d (%s): %s", e.Code, e.Status, e.Message)
}

// SendRequestError is a utility function that fills the given
// http.ResponseWriter with the provided error as json formatted. When calling
// this function one should directly return. It makes use of RequestedError by
// checking if the provided error matches this type.
func SendRequestError(err error, w http.ResponseWriter) {
	var outErr RequestError
	requestError, ok := err.(RequestError)
	if ok {
		outErr = requestError
	} else {
		outErr = NewInternalError(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	js, err := json.MarshalIndent(outErr, "", "  ")
	if err != nil {
		_, err = w.Write([]byte(fmt.Sprintf("GRAVE: failed to marshal error (%s) to json: (%s)", outErr.Error(), err.Error())))
		if err != nil {
			fmt.Println("GRAVE: could not write the GRAVE error: " + err.Error())
		}
		return
	}

	_, err = w.Write(js)
	if err != nil {
		log.Error("GRAVE: could not write the error: " + err.Error())
	}
}

// NewInternalError represents a 500 error
func NewInternalError(msg string) RequestError {
	return RequestError{
		Message: msg,
		Code:    500,
		Status:  "Internal Server Error",
	}
}

// NewNotOkError used when a 200 response is expected
func NewNotOkError(resp *http.Response) RequestError {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return NewInternalError("failed to read body while catching error: " + err.Error())
	}
	return RequestError{
		Message: "expected status code 200",
		Code:    resp.StatusCode,
		Status:  resp.Status,
		Data:    map[string]interface{}{"body": string(body)},
	}
}

// NewNotCreatedError used when a 201 response is expected
func NewNotCreatedError(resp *http.Response) RequestError {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return NewInternalError("failed to read body while catching error: " + err.Error())
	}
	return RequestError{
		Message: "expected status code 201",
		Code:    resp.StatusCode,
		Status:  resp.Status,
		Data:    map[string]interface{}{"body": string(body)},
	}
}

// NewNotAcceptedError used when a 202 response is expected
func NewNotAcceptedError(resp *http.Response) RequestError {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return NewInternalError("failed to read body while catching error: " + err.Error())
	}
	return RequestError{
		Message: "expected status code 202",
		Code:    resp.StatusCode,
		Status:  resp.Status,
		Data:    map[string]interface{}{"body": string(body)},
	}
}

// RedirectWithErrorFlash sets an error flash and redirect to the path. If there
// is an error during this process, raises a json error. One must return after
// calling this function.
func RedirectWithErrorFlash(path, message string, w http.ResponseWriter, r *http.Request, store sessions.Store) {
	log.Error(message)
	err := AddFlash(w, r, fmt.Sprintf("<pre>%s</pre>", message), store, Error)
	if err != nil {
		SendRequestError(xerrors.Errorf("failed to add error flash: %v", err), w)
		return
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
}

// RedirectWithWarningFlash sets a warning flash and redirect to the path. If
// there is an error during this process, raises a json error. One must return
// after calling this function.
func RedirectWithWarningFlash(path, message string, w http.ResponseWriter, r *http.Request, store sessions.Store) {
	err := AddFlash(w, r, fmt.Sprintf("%s", message), store, Warning)
	if err != nil {
		SendRequestError(xerrors.Errorf("failed to add warning flash: %v", err), w)
		return
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
}

// RedirectWithInfoFlash sets a warning flash and redirect to the path. If
// there is an error during this process, raises a json error. One must return
// after calling this function.
func RedirectWithInfoFlash(path, message string, w http.ResponseWriter, r *http.Request, store sessions.Store) {
	err := AddFlash(w, r, fmt.Sprintf("%s", message), store, Info)
	if err != nil {
		SendRequestError(xerrors.Errorf("failed to add info flash: %v", err), w)
		return
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
}
