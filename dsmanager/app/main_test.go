package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

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
