package main

import (
	"encoding/gob"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/dedis/odyssey/domanager/app/controllers"
	"github.com/dedis/odyssey/domanager/app/models"
	xhelpers "github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/require"
)

// I should be able to browse the home page without any problem
func TestHome_normal(t *testing.T) {
	gob.Register(xhelpers.Flash{})
	gob.Register(models.Session{})

	store := sessions.NewCookieStore([]byte("TOBECHANGEDOFCOURSE"))
	conf := &models.Config{}
	handler := controllers.HomeHandler(store, conf)

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
}

// It I am not logged I should not be able to go to the upload of a dataset or
// the list of uploaded datasets
func TestDataset_notLogged(t *testing.T) {
	gob.Register(xhelpers.Flash{})
	gob.Register(models.Session{})

	store := sessions.NewCookieStore([]byte("TOBECHANGEDOFCOURSE"))
	conf := &models.Config{}

	mux := http.NewServeMux()
	mux.Handle("/", controllers.HomeHandler(store, conf))
	mux.Handle("/datasets", controllers.DatasetIndexHandler(store, conf))

	server := httptest.NewServer(mux)
	defer server.Close()

	// telling the client to track the cookies
	// Thanks to https://www.meetspaceapp.com/2016/05/16/acceptance-testing-go-webapps-with-cookies.html
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{Jar: jar}

	req, err := http.NewRequest("GET", server.URL+"/datasets", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// We should have been redirected to home (ie. "")
	require.Equal(t, "", resp.Header.Get("Location"))
	// We should have a message telling us to login
	bodyBuf, err := ioutil.ReadAll(resp.Body)
	loginMessage := regexp.MustCompile("you need to be logged in to access this page")
	require.True(t, loginMessage.MatchString(string(bodyBuf)))
}

// An owner should be able to see its list of datasets
func TestDataset_index(t *testing.T) {

}
