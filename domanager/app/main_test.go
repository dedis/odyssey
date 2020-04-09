package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/dedis/odyssey/domanager/app/controllers"
	"github.com/dedis/odyssey/domanager/app/models"
	"github.com/dedis/odyssey/dsmanager/app/helpers"
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
	loginMessage := regexp.MustCompile("you need to be logged in to access this page")
	bodyBuf, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
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
	require.NoError(t, err)
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

	singinMessage := regexp.MustCompile("Please upload your credentials")
	bodyBuf, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
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

	loggedMessage := regexp.MustCompile("Identity file uploaded and set!")
	bodyBuf, err = ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
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
	require.NoError(t, err)
	require.True(t, uploadMsg.MatchString(string(bodyBuf)))
}

// If I upload a dataset I should be redirected to home and a new task should be
// created containing the different steps of the upload.
func Test_Dataset_POST(t *testing.T) {
	gob.Register(xhelpers.Flash{})
	gob.Register(models.Session{})

	taskManager := newFakeTaskManager()
	cloudClient := &fakeCloudClient{}

	store := sessions.NewCookieStore([]byte("TOBECHANGEDOFCOURSE"))
	conf := &models.Config{
		TOMLConfig:  &models.TOMLConfig{Standalone: true},
		TaskManager: taskManager,
		Executor:    fakeExecutor{},
		CloudClient: cloudClient,
	}

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

	taskCreadedMsg := regexp.MustCompile("Task to create dataset with index 0 created")
	bodyBuf, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.True(t, taskCreadedMsg.MatchString(string(bodyBuf)))

	select {
	case <-taskManager.called:
	case <-time.After(time.Second):
		t.Error("the taskmanager should have been called")
	}
	require.Equal(t, 1, len(taskManager.taskList))
	task := taskManager.taskList[0]

	timeout := time.Second
	var event xhelpers.TaskEvent

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DO Manager", event.Source)
	require.Equal(t, "starting the upload process", event.Message)
	require.Equal(t, "got this POST form: map[description:[dataset description] title:[dataset title]]", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DO Manager", event.Source)
	require.Equal(t, "Darc created in standalone mode", event.Message)
	require.Equal(t, "DarcID: darc:0000000000000000000000000000000000000000000000000000000000000000", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DO Manager", event.Source)
	require.Equal(t, "getting the dataset file", event.Message)
	require.Equal(t, "retrieving the 'dataset-file' post argument", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DO Manager", event.Source)
	require.Equal(t, "genering a symmetric key", event.Message)
	require.Equal(t, "genering a 16 bytes symmetric key", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DO Manager", event.Source)
	require.Equal(t, "genering an nonce", event.Message)
	require.Equal(t, "genering a 12 bytes nonce", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DO Manager", event.Source)
	require.Equal(t, "encrypting the dataset", event.Message)
	require.Equal(t, "encrypting with AES using the Galois Counter Mode", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DO Manager", event.Source)
	require.Equal(t, "uploading the encrypted dataset on the cloud", event.Message)
	require.True(t, strings.HasSuffix(event.Details, "_dataset+title.txt.aes"))

	// The cloud client should have been called
	require.True(t, cloudClient.called)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DO Manager", event.Source)
	require.Equal(t, "computing the SHA2", event.Message)
	require.Equal(t, "using the unencrypted file to compute the SHA2", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DO Manager", event.Source)
	require.Equal(t, "creating a Calypso write", event.Message)
	require.Equal(t, "using csadmin to create the calypso write that contains the symetric key and the nonce", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DO Manager", event.Source)
	require.Equal(t, "getting the write instance ID", event.Message)
	require.Equal(t, "parsing the output of csadmin to extract the write instance ID", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DO Manager", event.Source)
	require.Equal(t, "updating the catalog", event.Message)
	require.True(t, strings.HasPrefix(event.Details, "using the following command: [./catadmin"), "got this instead: %s", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DO Manager", event.Source)
	require.Equal(t, "dataset created", event.Message)

	select {
	case <-task.doneChan:
	case <-time.After(time.Second):
		t.Error("the task is not done after timeout")
	}
}

// -----------------------------------------------------------------------------
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

// Task manager

type fakeTaskManager struct {
	taskList []*FakeTask
	called   chan interface{}
}

func newFakeTaskManager() *fakeTaskManager {
	return &fakeTaskManager{
		taskList: make([]*FakeTask, 0),
		called:   make(chan interface{}),
	}
}

func (ftm *fakeTaskManager) NewTask(title string) xhelpers.TaskI {
	task := &FakeTask{
		eventChan: make(chan xhelpers.TaskEvent, 10),
		doneChan:  make(chan interface{}),
		data:      &helpers.TaskData{},
	}
	ftm.taskList = append(ftm.taskList, task)
	// So we can block in the test until at least one task is created
	close(ftm.called)
	return task
}

func (ftm *fakeTaskManager) NumTasks() int {
	return len(ftm.taskList)
}

func (ftm *fakeTaskManager) GetTask(index int) xhelpers.TaskI {
	return ftm.taskList[index]
}

func (ftm *fakeTaskManager) GetSortedTasks() []helpers.TaskI {
	return nil
}
func (ftm *fakeTaskManager) DeleteAllTasks() {

}
func (ftm *fakeTaskManager) RestoreTasks(tasks []helpers.TaskI) error {
	return nil
}

type FakeTask struct {
	eventChan chan xhelpers.TaskEvent
	doneChan  chan interface{}
	data      *helpers.TaskData
}

func (ft FakeTask) CloseError(source, msg, details string) {
	ft.eventChan <- xhelpers.NewTaskEventCloseError(source, msg, details)
	close(ft.doneChan)
}

func (ft FakeTask) AddInfo(source, msg, details string) {
	ft.eventChan <- xhelpers.NewTaskEvent(xhelpers.TypeInfo, source, msg, details)
}

func (ft FakeTask) AddInfof(source, msg, details string, args ...interface{}) {
	ft.eventChan <- xhelpers.NewTaskEvent(xhelpers.TypeInfo, source, msg, fmt.Sprintf(details, args...))
}

func (ft FakeTask) CloseOK(source, msg, details string) {
	ft.eventChan <- xhelpers.NewTaskEvent(xhelpers.TypeCloseOK, source, msg, details)
	close(ft.doneChan)
}

func (ft *FakeTask) AddTaskEvent(event helpers.TaskEvent) {

}

func (ft *FakeTask) Subscribe() *helpers.Subscriber {
	return nil
}

func (ft *FakeTask) GetData() *helpers.TaskData {
	return ft.data
}

func (ft *FakeTask) StatusImg() string {
	return ""
}

func (ft *FakeTask) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (ft *FakeTask) UnmarshalBinary(data []byte) error {
	return nil
}

// Executor

type fakeExecutor struct {
}

func (fe fakeExecutor) Run(args ...string) (bytes.Buffer, error) {
	cmdString := strings.Join(args, " ")
	// fmt.Printf("running: '%s'\n", cmdString)

	outb := bytes.Buffer{}

	if strings.HasPrefix(cmdString, "./bcadmin darc add") {
		outb.WriteString("darc:0000000000000000000000000000000000000000000000000000000000000000\n[]")
		return outb, nil
	}

	if strings.HasPrefix(cmdString, "./csadmin -c  contract write spawn ") {
		outb.WriteString("blabla\naaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	}

	return outb, nil
}

// Cloud Client

type fakeCloudClient struct {
	called bool
}

func (fcc *fakeCloudClient) PutObject(bucketName, objectName string, reader io.Reader, objectSize int64,
	opts interface{}) (n int64, err error) {
	fcc.called = true

	return 0, nil
}
