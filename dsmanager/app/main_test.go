package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/dedis/odyssey/dsmanager/app/controllers"
	"github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/dedis/odyssey/dsmanager/app/models"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/require"
)

// I should be able to go to the home page
func Test_Home(t *testing.T) {
	gob.Register(helpers.Flash{})

	store := sessions.NewCookieStore([]byte("TOBECHANGEDOFCOURSE"))
	conf := &models.Config{}
	handler := controllers.HomeHandler(store, conf)

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
}

// I should be able to see the list of projects
func Test_ProjectsGet(t *testing.T) {
	gob.Register(helpers.Flash{})

	store := sessions.NewCookieStore([]byte("TOBECHANGEDOFCOURSE"))
	conf := &models.Config{}
	handler := controllers.ProjectsIndexHandler(store, conf)

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/projects")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	projectTitle := regexp.MustCompile("<h1>List of Projects</h1>")
	bodyBuff, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.True(t, projectTitle.MatchString(string(bodyBuff)), "got: %s", bodyBuff)
}

// I should be able to see the list of datasets
func Test_Datasets(t *testing.T) {
	gob.Register(helpers.Flash{})

	store := sessions.NewCookieStore([]byte("TOBECHANGEDOFCOURSE"))
	conf := &models.Config{
		TOMLConfig: &models.TOMLConfig{},
		Executor:   fakeExecutor{},
	}
	handler := controllers.DatasetsIndexHandler(store, conf)

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/datasets")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	datasetsTitle := regexp.MustCompile("<h1>List of Datasets</h1>")
	bodyBuff, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.True(t, datasetsTitle.MatchString(string(bodyBuff)), "got: %s", bodyBuff)
}

// I should be able to create a new project (ie. initialize a new enclave) In
// this test we check only our part, ie. that we create the right task elements
// and that we call the enclave manager
func Test_Projects_POST(t *testing.T) {
	gob.Register(helpers.Flash{})

	taskManager := newFakeTaskManager()
	fakeRunHTTP := newFakeRunHTTP()

	store := sessions.NewCookieStore([]byte("TOBECHANGEDOFCOURSE"))
	conf := &models.Config{
		TOMLConfig: &models.TOMLConfig{
			// The read of the key could be abstracted with a reader manager
			PubKeyPath: "test/key.txt",
		},
		TaskManager: taskManager,
		Executor:    fakeExecutor{},
		RunHTTP:     fakeRunHTTP,
	}
	handler := controllers.ProjectsIndexHandler(store, conf)

	server := httptest.NewServer(handler)
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{Jar: jar}

	formData := url.Values{
		"datasetIDs": {"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}

	req, err := http.NewRequest("POST", server.URL+"/projects",
		strings.NewReader(formData.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	flashTitle := regexp.MustCompile("A new project with uniq id '.*' and title '.*' has been " +
		"created and the request to set up an enclave submitted")
	bodyBuff, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.True(t, flashTitle.MatchString(string(bodyBuff)), "got: %s", bodyBuff)

	select {
	case <-taskManager.called:
	case <-time.After(time.Second):
		t.Error("the taskmanager should have been called")
	}
	require.Equal(t, 1, len(taskManager.taskList))
	task := taskManager.taskList[0]

	timeout := time.Second
	var event helpers.TaskEvent

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DS Manager", event.Source)
	require.Equal(t, "task created", event.Message)
	require.Equal(t, "from the RequestCreateProjectInstance function", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DS Manager", event.Source)
	require.Equal(t, "Trying now to read the public key", event.Message)
	require.Equal(t, "public key: test/key.txt", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DS Manager", event.Source)
	require.Equal(t, "Project should be created", event.Message)
	require.Equal(t, "here is the output of the command:\n"+
		"Instance ID\nbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DS Manager", event.Source)
	require.Equal(t, "Got the right project instance id", event.Message)
	require.Equal(t, "Project instance ID: bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", event.Details)

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DS Manager", event.Source)
	require.Equal(t, "Sending a request to the enclave manager", event.Message)
	require.Equal(t, "POST localhost:5000/vapps", event.Details)

	select {
	case <-fakeRunHTTP.called:
	case <-time.After(time.Second):
		t.Error("the fakeRunHTTP should have been called")
	}

	select {
	case event = <-task.eventChan:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "DS Manager", event.Source)
	require.Equal(t, "Reading the status code", event.Message)
	require.Equal(t, "Status code: ", event.Details)

	// The task close is done by the enclave manager, this is not the our duty
	// to test it there
}

// -----------------------------------------------------------------------------
// Utility functions

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

func (ftm *fakeTaskManager) NewTask(title string) helpers.TaskI {
	task := &FakeTask{
		eventChan: make(chan helpers.TaskEvent, 10),
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

func (ftm *fakeTaskManager) GetTask(index int) helpers.TaskI {
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
	eventChan chan helpers.TaskEvent
	doneChan  chan interface{}
	data      *helpers.TaskData
}

func (ft FakeTask) CloseError(source, msg, details string) {
	ft.eventChan <- helpers.NewTaskEventCloseError(source, msg, details)
	close(ft.doneChan)
}

func (ft FakeTask) AddInfo(source, msg, details string) {
	ft.eventChan <- helpers.NewTaskEvent(helpers.TypeInfo, source, msg, details)
}

func (ft FakeTask) AddInfof(source, msg, details string, args ...interface{}) {
	ft.eventChan <- helpers.NewTaskEvent(helpers.TypeInfo, source, msg, fmt.Sprintf(details, args...))
}

func (ft FakeTask) CloseOK(source, msg, details string) {
	ft.eventChan <- helpers.NewTaskEvent(helpers.TypeCloseOK, source, msg, details)
	close(ft.doneChan)
}

func (ft *FakeTask) AddTaskEvent(event helpers.TaskEvent) {

}

func (ft *FakeTask) Subscribe() *helpers.Subscriber {
	sub := helpers.Subscriber{
		TaskStream: make(chan *helpers.TaskEvent),
	}
	return &sub
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

	if strings.HasPrefix(cmdString, "./pcadmin -c  contract project spawn") {
		outb.WriteString("Instance ID\nbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
		return outb, nil
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

// RunHTTP

type fakeRunHTTP struct {
	// Used to wait until it is called at least once
	called chan interface{}
}

func (f *fakeRunHTTP) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	close(f.called)
	return &http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer([]byte{})),
	}, nil
}

func newFakeRunHTTP() *fakeRunHTTP {
	return &fakeRunHTTP{
		called: make(chan interface{}),
	}
}
