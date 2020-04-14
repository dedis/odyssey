package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	xhelpers "github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/dedis/odyssey/enclavem/app/controllers"
	"github.com/dedis/odyssey/enclavem/app/models"
	"github.com/gorilla/sessions"
	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/onet/v3/log"
	"golang.org/x/xerrors"
)

// it should be able to contact the cloud manager to create a new vApp
func Test_Vapps_POST(t *testing.T) {
	gob.Register(xhelpers.Flash{})

	fakeRunHTTP := newFakeRunHTTP()

	store := sessions.NewCookieStore([]byte("TOBECHANGEDOFCOURSE"))
	conf := &models.Config{
		TOMLConfig:  &models.TOMLConfig{},
		Executor:    fakeExecutor{},
		RunHTTP:     fakeRunHTTP,
		CloudClient: &fakeCloudClient{},
	}
	handler := controllers.VappsIndexHandler(store, conf)

	server := httptest.NewServer(handler)
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{Jar: jar}

	formData := url.Values{
		"projectInstID": {"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		"projectUID":    {"TEST_UID"},
		"requestIndex":  {"99"},
	}

	req, err := http.NewRequest("POST", server.URL+"/projects",
		strings.NewReader(formData.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	reader := bufio.NewReader(resp.Body)

	responses := make(chan xhelpers.TaskEvent)

	// This routine will listen to http flush data and fill the responses chan
	go func() {
		for {
			line, err := reader.ReadBytes('\n')
			if err == io.EOF {
				log.Info("End of readloop")
				close(responses)
				return
			}
			if err != nil {
				t.Errorf("Response errored: %s", err.Error())
				return
			}

			if len(line) == 0 || string(line) == "\n" {
				continue
			}

			sections := bytes.SplitN(line, []byte(":"), 2)
			if len(sections) != 2 {
				t.Errorf("Malformatted data: %s", line)
				return
			}
			field, value := string(sections[0]), sections[1]
			switch field {
			case "data":
				taskResponse, err := processStreamValue(value)
				if err != nil {
					t.Errorf("Failed to read stream: %s", err.Error())
					return
				}
				responses <- *taskResponse
			default:
				t.Errorf("Unsupported field: %s", field)
				return
			}
		}
	}()

	timeout := time.Second
	var event xhelpers.TaskEvent

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "hi from the enclave manager", event.Message)
	require.Equal(t, "let's prpare this VM - from VappsIndexPost", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "parse arguments", event.Message)
	require.Equal(t, "Trying to parse the POST arguments", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "successfully got the project instance ID", event.Message)
	require.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "updating the contract status", event.Message)
	require.Equal(t, "setting 'preparing' on the project's instance status", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "getting argument", event.Message)
	require.Equal(t, "now let's try to get the project UID from the POST form", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "got the project UID", event.Message)
	require.Equal(t, "alright, we got this project UID: TEST_UID.This UID will be used as the "+
		"cloud endpoint and the enclave UID", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "getting the request index", event.Message)
	require.Equal(t, "now let's try to get the request index from the POST form", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "got the index request", event.Message)
	require.Equal(t, "alright, we got this request index: 99. This request index will be "+
		"used by the enclave to send logs at "+
		"/{projectUID}/logs/{requestIndex}/{timestamp}", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "trying to authenticate", event.Message)
	require.Equal(t, "trying to get an authentication token for vCloud", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "authentication OK", event.Message, event.Details)
	require.Equal(t, "we got the authentication token", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "getting minio client", event.Message)
	require.Equal(t, "creating the minio client based on the ENV variables", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "creating the bucket", event.Message)
	require.Equal(t, "now we are creating the bucket 'TEST_UID'", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "bucket not created", event.Message)
	require.Equal(t, "bucket not created because it is already created", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "creating vApp", event.Message)
	require.Equal(t, "creating new app via POST", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "sending POST request", event.Message)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "reading status code", event.Message)
	require.Equal(t, "INSTANTIATE_OK", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "waiting for the POST vapp task to succeed", event.Message)
	require.Equal(t, "http://example.com/TASK_PATH", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "starting to poll", event.Message)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "task succeded", event.Message, event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "Vm created", event.Message)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "setting property of vm", event.Message)
	require.Equal(t, "http://example.com/VM_PATH", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "sending PUT request to update properties", event.Message)
	require.Equal(t, "http://example.com/VM_PATH/productSections", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "reading status code", event.Message)
	require.Equal(t, "productSectionsOK", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "waiting for the PUT properties task to succeed", event.Message)
	require.Equal(t, "http://example.com/TASK_PATH", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "starting to poll", event.Message)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "task succeded", event.Message, event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "Property (cloud endpoint to use) on Vm set", event.Message)
	require.Equal(t, "http://example.com/TASK_PATH", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "updating the VM network adapter", event.Message)
	require.Equal(t, "http://example.com/VM_PATH", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "sending the PUT request to update the network adapter", event.Message)
	require.Equal(t, "http://example.com/VM_PATH/networkConnectionSection", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "reading status code", event.Message)
	require.Equal(t, "networkConnectionSectionOK", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "waiting for the PUT networkConnectionSection task to succeed", event.Message)
	require.Equal(t, "http://example.com/TASK_PATH", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "starting to poll", event.Message)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "task succeded", event.Message, event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "Network setting on Vm set", event.Message)
	require.Equal(t, "http://example.com/TASK_PATH", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "sending the GET request to fetch the IP", event.Message)
	require.Equal(t, "Request URL: http://example.com/VM_PATH/networkConnectionSection", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "reading status code", event.Message)
	require.Equal(t, "networkConnectionSectionOK", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "runing pcadmin", event.Message)
	require.Equal(t, "setting the pub key of the enclave on the contract running [./pcadmin -c  contract project invoke setURL -bc  -sign  -darc  -enclaveURL 1.1.1.1 -i aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa]", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "getting the newly created vApp", event.Message)
	require.Equal(t, "http://example.com/VAPP_PATH", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "reading status code", event.Message)
	require.Equal(t, "vappgetOK", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "powering on the newly created vApp", event.Message)
	require.Equal(t, "http://example.com/VAPP_PATH", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "sending the POST request to power on the vApp", event.Message)
	require.Equal(t, "http://example.com/VAPP_PATH/power/action/powerOn", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "reading status code", event.Message)
	require.Equal(t, "powerOnOK", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "waiting for the POST power on task to succeed", event.Message)
	require.Equal(t, "http://example.com/TASK_PATH", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "starting to poll", event.Message)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "task succeded", event.Message, event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "Vm powered on", event.Message)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "polling the endpoint that should be created by "+
		"the vApp", event.Message)
	require.Equal(t, "TEST_UID", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "bucket found", event.Message)
	require.Equal(t, "TEST_UID", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "checking the pub key", event.Message)
	require.Equal(t, "getting and checking the content of 'pub_key.txt'", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "'pub_key.txt' found", event.Message)
	require.Equal(t, "TEST_UID/pub_key.txt", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "found the pub key", event.Message)
	require.Equal(t, "ok, everything is allright (found 'ed25519:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc'), we can power off the enclave", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "running pcadmin", event.Message)
	require.Equal(t, "setting the pub key of the enclave on the contract running [./pcadmin -c  contract project invoke setEnclavePubKey -bc  -sign  -darc  -pubKey ed25519:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc -i aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa]", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "shut down the emclave", event.Message)
	require.Equal(t, "shut down (undeploy) the newly created vApp", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "sending the POST request to shutdown (undeploy) the vApp", event.Message)
	require.Equal(t, "http://example.com/VAPP_PATH/action/undeploy", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "reading status code", event.Message)
	require.Equal(t, "powerOffOK", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "waiting for the POST shutdown task to succeed", event.Message)
	require.Equal(t, "http://example.com/TASK_PATH", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "starting to poll", event.Message)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "task succeded", event.Message, event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "Vm shut down", event.Message, event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "updating the contract status", event.Message)
	require.Equal(t, "setting the contract instance status to 'preparedOK'", event.Details)

	select {
	case event = <-responses:
	case <-time.After(timeout):
		t.Error("event didn't come after timeout")
	}
	require.Equal(t, "enclave manager", event.Source)
	require.Equal(t, "enclave successfully set up", event.Message)
	require.Equal(t, "new enclave created, configured, booted, working fine and now powered off", event.Details)
}

// -----------------------------------------------------------------------------
// Utility functions

func processStreamValue(value []byte) (*xhelpers.TaskEvent, error) {
	taskResponse := &xhelpers.TaskEvent{}
	err := json.Unmarshal(value, taskResponse)
	if err != nil {
		return nil, errors.New("failed to decode json from stream: " + err.Error())
	}

	return taskResponse, nil
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

// RunHTTP

type fakeRunHTTP struct {
}

func (f *fakeRunHTTP) Do(client *http.Client, req *http.Request) (resp *http.Response, err error) {

	path := req.URL.Path

	if strings.HasSuffix(path, "api/org/") {

		return &http.Response{
			Body:       ioutil.NopCloser(bytes.NewBuffer([]byte{})),
			StatusCode: 200,
		}, nil

	} else if strings.HasSuffix(path, "api/sessions") {

		return &http.Response{
			Body:       ioutil.NopCloser(bytes.NewBuffer([]byte{})),
			StatusCode: 200,
			Header:     map[string][]string{"X-Vcloud-Authorization": {"TEST_TOKEN"}},
		}, nil

	} else if strings.HasSuffix(path, "action/instantiateVAppTemplate") {

		vapp := &models.VApp{}
		vapp.Tasks.Task.Href = "http://example.com/TASK_PATH"
		vapp.Children.VM.Href = "http://example.com/VM_PATH"
		vapp.Href = "http://example.com/VAPP_PATH"

		vappBuf, err := xml.Marshal(vapp)
		if err != nil {
			log.Fatalf("failed to marshal VAPP: %v", err)
		}

		return &http.Response{
			Body:       ioutil.NopCloser(bytes.NewBuffer(vappBuf)),
			Status:     "INSTANTIATE_OK",
			StatusCode: 201,
		}, nil

	} else if path == "/TASK_PATH" {

		task := &models.Task{}
		task.Status = "success"
		taskBuf, err := xml.Marshal(task)
		if err != nil {
			log.Fatalf("failed to marshal Task: %v", err)
		}

		return &http.Response{
			Body:       ioutil.NopCloser(bytes.NewBuffer(taskBuf)),
			StatusCode: 200,
		}, nil

	} else if strings.HasSuffix(path, "/productSections") {

		task := &models.Task{}
		task.Status = "success"
		task.Href = "http://example.com/TASK_PATH"
		taskBuf, err := xml.Marshal(task)
		if err != nil {
			log.Fatalf("failed to marshal Task: %v", err)
		}

		return &http.Response{
			Body:       ioutil.NopCloser(bytes.NewBuffer(taskBuf)),
			StatusCode: 202,
			Status:     "productSectionsOK",
		}, nil

	} else if path == "/VM_PATH/networkConnectionSection" && req.Method == http.MethodPut {

		task := &models.Task{}
		task.Status = "success"
		task.Href = "http://example.com/TASK_PATH"
		taskBuf, err := xml.Marshal(task)
		if err != nil {
			log.Fatalf("failed to marshal Task: %v", err)
		}

		return &http.Response{
			Body:       ioutil.NopCloser(bytes.NewBuffer(taskBuf)),
			StatusCode: 202,
			Status:     "networkConnectionSectionOK",
		}, nil

	} else if path == "/VM_PATH/networkConnectionSection" && req.Method == http.MethodGet {

		ncs := &models.NetworkConnectionSection{}
		ncs.NetworkConnection.IPAddress = "1.1.1.1"
		ncsBuf, err := xml.Marshal(ncs)
		if err != nil {
			log.Fatalf("failed to marshal Task: %v", err)
		}

		return &http.Response{
			Body:       ioutil.NopCloser(bytes.NewBuffer(ncsBuf)),
			StatusCode: 200,
			Status:     "networkConnectionSectionOK",
		}, nil

	} else if path == "/VAPP_PATH" {

		vapp := &models.VApp{}
		vapp.Tasks.Task.Href = "http://example.com/TASK_PATH"
		vapp.Children.VM.Href = "http://example.com/VM_PATH"
		vapp.Href = "http://example.com/VAPP_PATH"
		vapp.Status = "8"

		vappBuf, err := xml.Marshal(vapp)
		if err != nil {
			log.Fatalf("failed to marshal VAPP: %v", err)
		}

		return &http.Response{
			Body:       ioutil.NopCloser(bytes.NewBuffer(vappBuf)),
			StatusCode: 200,
			Status:     "vappgetOK",
		}, nil

	} else if path == "/VAPP_PATH/power/action/powerOn" {

		task := &models.Task{}
		task.Status = "success"
		task.Href = "http://example.com/TASK_PATH"
		taskBuf, err := xml.Marshal(task)
		if err != nil {
			log.Fatalf("failed to marshal Task: %v", err)
		}

		return &http.Response{
			Body:       ioutil.NopCloser(bytes.NewBuffer(taskBuf)),
			StatusCode: 202,
			Status:     "powerOnOK",
		}, nil

	} else if path == "/VAPP_PATH/action/undeploy" {
		task := &models.Task{}
		task.Status = "success"
		task.Href = "http://example.com/TASK_PATH"
		taskBuf, err := xml.Marshal(task)
		if err != nil {
			log.Fatalf("failed to marshal Task: %v", err)
		}

		return &http.Response{
			Body:       ioutil.NopCloser(bytes.NewBuffer(taskBuf)),
			StatusCode: 202,
			Status:     "powerOffOK",
		}, nil
	}

	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBuffer([]byte{})),
		StatusCode: 200,
		Status:     "DEFAULT_STATUS",
	}, nil
}

func newFakeRunHTTP() *fakeRunHTTP {
	return &fakeRunHTTP{}
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

// GetObject gets an object
func (fcc *fakeCloudClient) GetObject(bucketName, objectName string,
	opts interface{}) (xhelpers.CloudObject, error) {

	if bucketName == "TEST_UID" {
		return &fakeCloudObject{}, nil
	}

	return nil, xerrors.New("no implemented")
}

// BucketExists tells if a bucket exists
func (fcc *fakeCloudClient) BucketExists(bucketName string) (bool, error) {
	return true, nil
}

// MakeBucket makes a bucket
func (fcc *fakeCloudClient) MakeBucket(bucketName string, location string) (err error) {
	return xerrors.New("no implemented")
}

type fakeCloudObject struct {
}

func (fco fakeCloudObject) Close() (err error) {
	return nil
}
func (fco fakeCloudObject) Stat() (minio.ObjectInfo, error) {
	return minio.ObjectInfo{}, nil
}
func (fco fakeCloudObject) Read(p []byte) (n int, err error) {
	pubKeyBuf := []byte("ed25519:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
	for i := 0; i < len(pubKeyBuf); i++ {
		p[i] = pubKeyBuf[i]
	}
	return len(pubKeyBuf), io.EOF
}
