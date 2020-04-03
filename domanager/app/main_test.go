package main

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/dedis/odyssey/domanager/app/controllers"
	"github.com/dedis/odyssey/domanager/app/models"
	xhelpers "github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/cothority/v3/byzcoin/bcadmin/lib"
)

// I should be able to browse the home page without any problem
func Test_Home(t *testing.T) {
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

// If I am not logged I should not be able to go to the upload of a dataset or
// the list of uploaded datasets
func Test_NotLogged(t *testing.T) {
	gob.Register(xhelpers.Flash{})
	gob.Register(models.Session{})

	store := sessions.NewCookieStore([]byte("TOBECHANGEDOFCOURSE"))
	conf := &models.Config{}

	mux := http.NewServeMux()
	mux.Handle("/", controllers.HomeHandler(store, conf))
	mux.Handle("/datasets", controllers.DatasetIndexHandler(store, conf))
	mux.Handle("/datasets/new", controllers.DatasetNewHandler(store, conf))

	server := httptest.NewServer(mux)
	defer server.Close()

	// telling the client to track the cookies
	// Thanks to https://www.meetspaceapp.com/2016/05/16/acceptance-testing-go-webapps-with-cookies.html
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{Jar: jar}

	// Checking /datasets
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

	// Checking /datasets/new
	req, err = http.NewRequest("GET", server.URL+"/datasets/new", nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// We should have been redirected to home (ie. "")
	require.Equal(t, "", resp.Header.Get("Location"))
	// We should have a message telling us to login
	bodyBuf, err = ioutil.ReadAll(resp.Body)
	require.True(t, loginMessage.MatchString(string(bodyBuf)))
}

// I should be able to login
func Test_Login(t *testing.T) {
	gob.Register(xhelpers.Flash{})
	gob.Register(models.Session{})

	store := sessions.NewCookieStore([]byte("TOBECHANGEDOFCOURSE"))
	conf := &models.Config{}

	mux := http.NewServeMux()
	mux.Handle("/", controllers.HomeHandler(store, conf))
	mux.Handle("/signin", controllers.SessionHandler(store, conf))

	server := httptest.NewServer(mux)
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{Jar: jar}

	// Checking /signin
	req, err := http.NewRequest("GET", server.URL+"/signin", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	singinMessage := regexp.MustCompile("Please upload your credentials")
	require.True(t, singinMessage.MatchString(string(bodyBuf)))

	// Creating the POST form to login
	file, err := os.Open("test/bc-test.cfg")
	require.NoError(t, err)
	fileContents, err := ioutil.ReadAll(file)
	require.NoError(t, err)
	fi, err := file.Stat()
	require.NoError(t, err)
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("myFile", fi.Name())
	require.NoError(t, err)
	part.Write(fileContents)
	err = writer.Close()
	require.NoError(t, err)

	// Sending a POST request to signin
	req, err = http.NewRequest("POST", server.URL+"/signin", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	bodyBuf, err = ioutil.ReadAll(resp.Body)
	loggedMessage := regexp.MustCompile("Identity file uploaded and set!")
	require.True(t, loggedMessage.MatchString(string(bodyBuf)))
}

// If I am logged I should be able to see the dataset upload page
func Test_Dataset_new(t *testing.T) {
	gob.Register(xhelpers.Flash{})
	gob.Register(models.Session{})

	store := sessions.NewCookieStore([]byte("TOBECHANGEDOFCOURSE"))
	conf := &models.Config{}

	mux := http.NewServeMux()
	mux.Handle("/", controllers.HomeHandler(store, conf))
	mux.Handle("/datasets/new", controllers.DatasetNewHandler(store, conf))

	server := httptest.NewServer(mux)
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{Jar: jar}

	cookie := setSession(t, server.URL)

	// Checking /datasets/new while we are logged
	req, err := http.NewRequest("GET", server.URL+"/datasets/new", nil)
	require.NoError(t, err)
	req.Header.Add("Cookie", cookie)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	uploadMsg := regexp.MustCompile("<h1>Upload a new dataset</h1>")
	bodyBuf, err := ioutil.ReadAll(resp.Body)
	require.True(t, uploadMsg.MatchString(string(bodyBuf)))
}

// If I upload a dataset I should be redirected to home and a new task should be
// created containing the different steps of the upload.
func Test_Dataset_POST(t *testing.T) {
	gob.Register(xhelpers.Flash{})
	gob.Register(models.Session{})

	store := sessions.NewCookieStore([]byte("TOBECHANGEDOFCOURSE"))
	conf := &models.Config{&models.TOMLConfig{}, xhelpers.DefaultTaskFactory{}}

	mux := mux.NewRouter()
	mux.Handle("/showtasks/{id}", controllers.ShowtasksShowHandler(store, conf))
	mux.Handle("/datasets", controllers.DatasetIndexHandler(store, conf))

	server := httptest.NewServer(mux)
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{Jar: jar}

	// signin
	cookie := setSession(t, server.URL)

	// Creating the POST form to login
	file, err := os.Open("test/dataset.txt")
	require.NoError(t, err)
	fileContents, err := ioutil.ReadAll(file)
	require.NoError(t, err)
	fi, err := file.Stat()
	require.NoError(t, err)
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("dataset-file", fi.Name())
	require.NoError(t, err)
	part.Write(fileContents)
	err = writer.WriteField("title", "dataset title")
	require.NoError(t, err)
	err = writer.WriteField("description", "dataset description")
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	// Preparing the request
	req, err := http.NewRequest("POST", server.URL+"/datasets", body)
	require.NoError(t, err)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Cookie", cookie)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	bodyBuf, err := ioutil.ReadAll(resp.Body)
	taskCreadedMsg := regexp.MustCompile("Task to create dataset with index 0 created")
	require.True(t, taskCreadedMsg.MatchString(string(bodyBuf)))
}

// -----------------
// Utility functions

// setSession sets the session variable that creates the session and returns the
// cookie that the client must use.
func setSession(t *testing.T, url string) string {
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	session, err := store.Get(req, "signin-session")
	randKey := lib.RandString(12)
	session.Values["key"] = randKey
	bcPath := "test/bc-test.cfg"
	cfg, _, err := lib.LoadConfig(bcPath)
	require.NoError(t, err)
	models.SaveSession(randKey, bcPath, &cfg)
	err = session.Save(req, w)
	require.NoError(t, err)

	cookies, ok := w.Header()["Set-Cookie"]
	require.True(t, ok)
	require.Equal(t, 1, len(cookies))

	return cookies[0]
}
