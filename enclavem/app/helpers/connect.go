package helpers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

/*
	This file handles the connection to the vCloud Director API.
*/

var (
	// Is set at the first connection and renewed if a connection fails
	authToken = ""
	// VcdAPIPass is the Vcloud Director API password
	VcdAPIPass = os.Getenv("VCD_PASS")
	// VcdHost is the Vcloud Director host
	VcdHost = os.Getenv("VCD_HOST")
	// VcdAPIUser is the Vcloud Director API user
	VcdAPIUser = os.Getenv("VCD_API_USER")
	// VcdOrgName is the Vcloud Director organization name
	VcdOrgName = os.Getenv("VCD_ORG_NAME")
)

// GetToken checks if a token is already set and, if not, request one. If a
// token is already sets, it checks if the token is still valid by sending a GET
// to '/org/'. If the request doesn't work, a new token is requested.
// It then finally return a working token.
func GetToken(w http.ResponseWriter) (string, error) {
	return getToken(w, true)
}

func getToken(w http.ResponseWriter, retry bool) (string, error) {
	if authToken == "" {
		err := connect(w)
		if err != nil {
			return "", NewInternalError("failed to connect: " + err.Error())
		}
	}
	url := fmt.Sprintf("https://%s/api/org/", VcdHost)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", NewInternalError("failed to build request:" + err.Error())
	}
	req.Header.Set("X-Vcloud-Authorization", authToken)
	req.Header.Set("Accept", "application/*;version=31.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", NewInternalError(fmt.Sprintf("failed to get %s: %s", url, err.Error()))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if retry {
			// The token might be expired, so we should try to get a new one
			authToken = ""
			return getToken(w, false)
		}
		return "", fmt.Errorf("failed to check the token at %s, expected 200 status code but got %s", url, resp.Status)
	}

	return authToken, nil
}

// connect gets and sets a new access token from vCloud
func connect(w http.ResponseWriter) error {
	url := fmt.Sprintf("https://%s/api/sessions", VcdHost)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return errors.New("failed to build request:" + err.Error())
	}
	req.SetBasicAuth(VcdAPIUser, VcdAPIPass)
	req.Header.Set("Accept", "application/*;version=31.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("failed to send request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return NewNotOkError(resp)
	}

	headerAuth, ok := resp.Header["X-Vcloud-Authorization"]
	if !ok {
		return errors.New("failed to get the X-Vcloud-Authorization header, its not there")
	}
	if len(headerAuth) == 0 {
		return errors.New("len of X-Vcloud-Authorization is 0")
	}
	if len(headerAuth[0]) == 0 {
		return errors.New("failed to get the authToken, it's empty")
	}

	authToken = headerAuth[0]

	return nil
}
