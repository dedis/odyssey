package models

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dedis/odyssey/catalogc"
	"github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/dedis/odyssey/projectc"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/protobuf"
	"golang.org/x/xerrors"
)

// ProjectList holds all the project created
var ProjectList = make(map[string]*Project)

// ProjectStatus indicates the status of a project
type ProjectStatus string

// RequestStatus indicates the status of a request
type RequestStatus string

const (
	// ProjectStatusInit indicates that the project was created but with nothing
	// more
	ProjectStatusInit = "initialized"
	// ProjectStatusPreparingEnclave indicates that a request to boot the enclave
	// has been sent
	ProjectStatusPreparingEnclave = "preparingEnclave"
	// ProjectStatusPreparingEnclaveErrored indicates that a request to boot the
	// enclave has been sent and that is errored
	ProjectStatusPreparingEnclaveErrored = "preparingEnclaveErrored"
	// ProjectStatusPreparingEnclaveDone indicates that a request to boot the
	// enclave has been sent and that it terminated well
	ProjectStatusPreparingEnclaveDone = "preparingEnclaveDone"
	// ProjectStatusUpdatingAttributes indicates that a request to update the
	// attributes has been submitted
	ProjectStatusUpdatingAttributes = "updatingAttributes"
	// ProjectStatusAttributesUpdatedDone indicated the attributes have been updated
	ProjectStatusAttributesUpdatedDone = "attributesUpdated"
	// ProjectStatusAttributesUpdatedErrored indicates an error occured while
	// updating the attributes.
	ProjectStatusAttributesUpdatedErrored = "attributesUpdatedErrored"
	// ProjectStatusUnlockingEnclave indicates that a request to unlock the
	// enclave has been submitted
	ProjectStatusUnlockingEnclave = "unlockingEnclave"
	// ProjectStatusUnlockingEnclaveDone indicates that a request to unlock the
	// enclave has been submitted and is done
	ProjectStatusUnlockingEnclaveDone = "unlockingEnclaveDone"
	// ProjectStatusUnlockingEnclaveErrored indicates that a request to unlock the
	// enclave has been submitted and errored
	ProjectStatusUnlockingEnclaveErrored = "unlockingEnclaveErrored"
	// ProjectStatusDeletingEnclave indicate that we are trying to delete the
	// enclave
	ProjectStatusDeletingEnclave = "deletingEnclave"
	// ProjectStatusDeletingEnclaveDone indicate that the enclave was
	// successfully deleted
	ProjectStatusDeletingEnclaveDone = "deletingEnclaveDone"
	// ProjectStatusDeletingEnclaveErrored indicate that we failed to delete the
	// enclave
	ProjectStatusDeletingEnclaveErrored = "deletingEnclaveErrored"
	// RequestStatusRunning indicates the request is running
	RequestStatusRunning = "requestRunning"
	// RequestStatusErrored indicates the request stopped with an error
	RequestStatusErrored = "requestErrored"
	// RequestStatusDone indicates the request ended well
	RequestStatusDone = "requestDone"
)

// Project holds a project request
type Project struct {
	Status ProjectStatus
	// Uniq ID. Also servers as the enclave name and cloud endpoint.
	UID            string
	Title          string
	Description    string
	InstanceID     string
	Requests       []*Request
	StatusNotifier *helpers.StatusNotifier
	CreatedAt      time.Time
	PubKey         string
}

// Request holds a request for a project
type Request struct {
	Description    string
	Status         RequestStatus
	Tasks          []helpers.TaskI
	StatusNotifier *helpers.StatusNotifier
	Index          int
}

// RequestSorter is used to define a sorter that sorts requests by the Index
// field.
type RequestSorter []*Request

// Len returns the len ...
func (p RequestSorter) Len() int { return len(p) }

// Swap swaps ...
func (p RequestSorter) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Less compares two projects based on their CreatedAt fields
func (p RequestSorter) Less(i, j int) bool {
	return p[i].Index < p[j].Index
}

// GetCloudAttributes return the alias, bucket name and prefix
func (r *Request) GetCloudAttributes(projectUID string) (string, string, string) {
	return "dedis", projectUID, fmt.Sprintf("logs/%d", r.Index)
}

// ProjectSorter is used to define a sorter that sorts projects by the CreatedAt
// field.
type ProjectSorter []*Project

// Len returns the len ...
func (p ProjectSorter) Len() int { return len(p) }

// Swap swaps ...
func (p ProjectSorter) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Less compares two projects based on their CreatedAt fields
func (p ProjectSorter) Less(i, j int) bool {
	return p[i].CreatedAt.Before(p[j].CreatedAt)
}

// NewProject creates an empty project. If the title is empty (""), will use a
// random generated title.
func NewProject(title, description string) *Project {
	timePrefix := time.Now().Format("2006-01-02-1504-05")
	endpointBuf := make([]byte, 4)
	// The error is not checked here. We tolerate the endpointBuf beeing set at
	// zero.
	rand.Read(endpointBuf)
	idStr := fmt.Sprintf("%s-%x", timePrefix, endpointBuf)
	if title == "" {
		title = helpers.GetRandomName()
	}
	project := &Project{
		Status:         ProjectStatusInit,
		UID:            idStr,
		Title:          title,
		Description:    description,
		Requests:       make([]*Request, 0),
		StatusNotifier: helpers.NewStatusNotifier(),
		CreatedAt:      time.Now(),
	}
	ProjectList[idStr] = project
	return project
}

// AddRequest adds a new request to the project by prepending the request to the
// list of existing requests.
func (p *Project) AddRequest(request *Request) {
	request.Index = len(p.Requests)
	p.Requests = append(p.Requests, request)
}

// PrepareBeforeMarshal updates the elements of the project that can not be
// marshalled.
func (p *Project) PrepareBeforeMarshal() {
	p.StatusNotifier = nil
	for _, request := range p.Requests {
		request.StatusNotifier = nil
		for _, task := range request.Tasks {
			task.SetSubscribers(nil)
		}
	}
}

// PrepareAfterUnmarshal updates the elements of the project that were changed
// before marshalling.
func (p *Project) PrepareAfterUnmarshal() {
	p.StatusNotifier = helpers.NewStatusNotifier()
	p.StatusNotifier.Terminated = true
	for _, request := range p.Requests {
		request.StatusNotifier = helpers.NewStatusNotifier()
		request.StatusNotifier.Terminated = true
		for _, task := range request.Tasks {
			task.SetSubscribers(make([]*helpers.Subscriber, 0))
		}
	}
}

// GetLastestTaskMsg return the latest task message, or an empty one if there
// isn't any. This is convenient to check if the last message was an error and
// display an output to the client, for example the reasons why the enclave
// can't be unlocked.
func (p *Project) GetLastestTaskMsg() (string, string) {
	if len(p.Requests) == 0 {
		return "", ""
	}
	r := p.Requests[len(p.Requests)-1]
	if len(r.Tasks) == 0 {
		return "", ""
	}
	t := r.Tasks[len(r.Tasks)-1]
	if len(t.GetHistory()) == 0 {
		return "", ""
	}
	return t.GetHistory()[0].Message, t.GetHistory()[0].Details
}

// RequestCreateProjectInstance creates and runs a request that creates a new
// instance of a project contract.
func (p *Project) RequestCreateProjectInstance(datasetIDs []string, conf *Config) (*Request, helpers.TaskI) {
	p.Status = ProjectStatusPreparingEnclave
	p.StatusNotifier.UpdateStatus(ProjectStatusPreparingEnclave)

	task := conf.TaskManager.NewTask("New project creation")
	tef := helpers.NewTaskEventFactory("ds manager")

	// We use this client to listen to the events and update the project
	// status based on what we receive.
	client := task.Subscribe()

	task.AddInfo(tef.Source, "task created", "from the RequestCreateProjectInstance function")

	request := &Request{
		Description:    "Prepare the enclave",
		Tasks:          []helpers.TaskI{task},
		StatusNotifier: helpers.NewStatusNotifier(),
		Status:         RequestStatusRunning,
	}
	p.AddRequest(request)

	go func() {
		for {
			select {
			case taskEl := <-client.TaskStream:
				if taskEl.Type == helpers.TypeCloseError {

					p.Status = ProjectStatusPreparingEnclaveErrored
					p.StatusNotifier.UpdateStatusAndClose(
						ProjectStatusPreparingEnclaveErrored)

					request.Status = RequestStatusErrored
					request.StatusNotifier.UpdateStatusAndClose(
						RequestStatusErrored)

				} else if taskEl.Type == helpers.TypeCloseOK {

					p.Status = ProjectStatusPreparingEnclaveDone
					p.StatusNotifier.UpdateStatusAndClose(
						ProjectStatusPreparingEnclaveDone)

					request.Status = RequestStatusDone
					request.StatusNotifier.UpdateStatusAndClose(
						RequestStatusDone)
				}
			default:
				select {
				case <-client.Done:
					return
				default:
				}
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	datasetR, err := regexp.Compile("^[0-9a-f]{64}$")
	if err != nil {
		task.CloseError(tef.Source, "failed to compile regex", err.Error())
		return nil, nil
	}

	for _, datasetID := range datasetIDs {
		ok := datasetR.MatchString(datasetID)
		if !ok {
			task.CloseError(tef.Source, "got a wong dataset ID", "dataset ID foun: "+datasetID)
			return nil, nil
		}
	}
	idStr := strings.Join(datasetIDs, ",")

	task.AddInfo(tef.Source, "Trying now to read the public key", "public key: "+conf.PubKeyPath)

	pubKeyBuf, err := ioutil.ReadFile(conf.PubKeyPath)
	if err != nil {
		task.CloseError(tef.Source, "Failed to read pub key.", err.Error())
		return nil, nil
	}

	pubKeyStr := string(pubKeyBuf)
	log.Lvl1("Here is the pubKey: ", pubKeyStr)

	p.PubKey = pubKeyStr

	output, err := createProjectInstace(idStr, pubKeyStr, conf)
	if err != nil {
		task.CloseError(tef.Source, "failed to create project instance", err.Error())
		return nil, nil
	}
	task.AddInfo(tef.Source, "Project should be created", "here is the output of the command:\n"+output)

	// We know the bcadmin command will output the instID at the second line
	outputSplit := strings.Split(output, "\n")
	if len(outputSplit) < 2 {
		task.CloseError(tef.Source, "Got unexpected output split",
			fmt.Sprintf("Here is the output split: %s", outputSplit))
		return nil, nil
	}

	projectInstID := outputSplit[1]
	ok := datasetR.MatchString(projectInstID)
	if !ok {
		log.Info("got a wong project ID: " + projectInstID)
		task.CloseError(tef.Source, "got a wong project ID", "ProjectInstID: "+projectInstID)
		return nil, nil
	}

	p.InstanceID = projectInstID
	task.AddInfo(tef.Source, "Got the right project instance id", "Project instance ID: "+p.InstanceID)

	return request, task
}

// RequestBootEnclave talks to the enclave manager and asks it to boot an
// enclave.
func (p *Project) RequestBootEnclave(request *Request, task helpers.TaskI, conf *Config) {
	tef := helpers.NewTaskEventFactory("ds manager")

	// This is the case where RequestCreateProjectInstance has not been called
	// before because we are trying to re-request to boot the enclave, but the
	// project instance has already been created with a previous request.
	if request == nil || task == nil {
		p.Status = ProjectStatusPreparingEnclave
		p.StatusNotifier.UpdateStatus(ProjectStatusPreparingEnclave)

		task = conf.TaskManager.NewTask("New project creation")

		// We use this client to listen to the events and update the project
		// status based on what we receive.
		client := task.Subscribe()

		task.AddInfo(tef.Source, "task created", "we are in the RequestBootEnclave function")

		request = &Request{
			Description:    "Prepare the enclave",
			Tasks:          []helpers.TaskI{task},
			StatusNotifier: helpers.NewStatusNotifier(),
			Status:         RequestStatusRunning,
		}
		p.AddRequest(request)

		go func() {
			for {
				select {
				case taskEl := <-client.TaskStream:
					if taskEl.Type == helpers.TypeCloseError {

						p.Status = ProjectStatusPreparingEnclaveErrored
						p.StatusNotifier.UpdateStatusAndClose(
							ProjectStatusPreparingEnclaveErrored)

						request.Status = RequestStatusErrored
						request.StatusNotifier.UpdateStatusAndClose(
							RequestStatusErrored)

					} else if taskEl.Type == helpers.TypeCloseOK {

						p.Status = ProjectStatusPreparingEnclaveDone
						p.StatusNotifier.UpdateStatusAndClose(
							ProjectStatusPreparingEnclaveDone)

						request.Status = RequestStatusDone
						request.StatusNotifier.UpdateStatusAndClose(
							RequestStatusDone)

					}
				default:
					select {
					case <-client.Done:
						return
					default:
					}
				}
				time.Sleep(time.Millisecond * 100)
			}
		}()
	}

	task.AddInfo(tef.Source, "Sending a request to the enclave manager", "POST localhost:5000/vapps")

	formData := url.Values{
		"projectInstID": {p.InstanceID},
		"projectUID":    {p.UID},
		"requestIndex":  {strconv.Itoa(request.Index)},
	}
	resp, err := http.PostForm("http://localhost:5000/vapps", formData)
	if err != nil {
		log.Infof("Failed to send request: %s", err.Error())
		task.CloseError(tef.Source, "Failed to send request", err.Error())
		return
	}
	defer resp.Body.Close()

	task.AddInfo(tef.Source, "Reading the status code", "Status code: "+resp.Status)

	if resp.StatusCode != 200 {
		log.Infof("Got an unexpected response: %s", resp.Status)
		task.CloseError(tef.Source, "Got an unexpected response", "Response: "+resp.Status)
		return
	}

	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			log.Info("End of readloop")
			return
		}
		if err != nil {
			log.Infof("Response errored: %s", err.Error())
			task.CloseError(tef.Source, "Response errored", err.Error())
			return
		}

		if len(line) == 0 || string(line) == "\n" {
			continue
		}
		log.Infof("here is the task: %s", line)

		sections := bytes.SplitN(line, []byte(":"), 2)
		if len(sections) != 2 {
			log.Infof("Malformatted data: %s", line)
			task.CloseError(tef.Source, "Malformatted data", "Here is the line: "+string(line))
			return
		}
		field, value := string(sections[0]), sections[1]
		switch field {
		case "data":
			taskResponse, err := processStreamValue(value)
			if err != nil {
				log.Infof("Failed to read stream: %s", err.Error())
				task.CloseError(tef.Source, "Failed to read stream", err.Error())
				return
			}
			task.AddTaskEvent(*taskResponse)
		default:
			log.Infof("Unsupported field: %s", field)
			task.CloseError(tef.Source, "Unsupported field", "Fiedl: "+field)
			return
		}
	}
}

func createProjectInstace(idStr, pubKey string, conf *Config) (string, error) {

	cmd := exec.Command("./pcadmin", "-c", conf.ConfigPath, "contract",
		"project", "spawn", "-is", idStr, "-bc", conf.BCPath, "-sign",
		conf.KeyID, "-darc", conf.DarcID, "-pubKey", pubKey)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run the command: %s - "+
			"Output: %s - Err: %s", err.Error(), outb.String(), errb.String())
	}

	cmdOut := outb.String()
	log.LLvl1("here is the output of the value contract:", cmdOut)
	return cmdOut, nil
}

func processStreamValue(value []byte) (*helpers.TaskEvent, error) {
	taskResponse := &helpers.TaskEvent{}
	err := json.Unmarshal(value, taskResponse)
	if err != nil {
		return nil, errors.New("failed to decode json from stream: " + err.Error())
	}

	return taskResponse, nil
}

// RequestUpdateAttributes creates a new request to update the attributes
func (p *Project) RequestUpdateAttributes(values url.Values, conf *Config) {
	tef := helpers.NewTaskEventFactory("ds manager")

	p.Status = ProjectStatusUpdatingAttributes
	p.StatusNotifier.Terminated = false
	p.StatusNotifier.UpdateStatus(ProjectStatusUpdatingAttributes)

	task := conf.TaskManager.NewTask("Update the attributes")

	// We use this client to listen to the events and update the project
	// status based on what we receive.
	client := task.Subscribe()

	task.AddInfo(tef.Source, "task created", "From the RequestUpdateAttributes function")

	request := &Request{
		Description:    "Update the enclave's attributes",
		Tasks:          []helpers.TaskI{task},
		StatusNotifier: helpers.NewStatusNotifier(),
		Status:         RequestStatusRunning,
	}
	p.AddRequest(request)

	go func() {
		for {
			select {
			case taskEl := <-client.TaskStream:
				if taskEl.Type == helpers.TypeCloseError {

					err := p.UpdateProjectcStatus(conf, "updatedAttrErrored")
					if err != nil {
						log.Error("failed to update contract status: ", err)
					}

					p.Status = ProjectStatusAttributesUpdatedErrored
					p.StatusNotifier.UpdateStatusAndClose(
						ProjectStatusAttributesUpdatedErrored)

					request.Status = RequestStatusErrored
					request.StatusNotifier.UpdateStatusAndClose(
						RequestStatusErrored)

				} else if taskEl.Type == helpers.TypeCloseOK {

					err := p.UpdateProjectcStatus(conf, "updatedAttrOK")
					if err != nil {
						log.Error("failed to update contract status: ", err)
					}

					p.Status = ProjectStatusAttributesUpdatedDone
					p.StatusNotifier.UpdateStatusAndClose(
						ProjectStatusAttributesUpdatedDone)

					request.Status = RequestStatusDone
					request.StatusNotifier.UpdateStatusAndClose(
						RequestStatusDone)
				}
			default:
				select {
				case <-client.Done:
					return
				default:
				}
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	err := p.UpdateProjectcStatus(conf, "updatingAttr")
	if err != nil {
		task.CloseError(tef.Source, "failed to update the projectc status",
			err.Error())
		return
	}

	// We get the scaffold metadata from the catalog, then we fill it up with
	// the given element from the HTML form. If an element is not found in the
	// scaffold form, we then create a new attribute. This is the case of
	// dataset specific attributes.
	cmd := exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
		"catalog", "getMetadata", "-i", conf.CatalogID, "-bc", conf.BCPath,
		"--export")

	task.AddInfo(tef.Source, "command created", fmt.Sprintf("%s", cmd.Args))
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	// We need to sleep because we just updated the status of projectc, which is
	// why we need to wait in order to have the right counter, otherwise we
	// could have an error of type "got counter=521, but need 522"
	time.Sleep(5 * time.Second)
	err = cmd.Run()
	if err != nil {
		task.CloseError(tef.Source, "failed to get the scaffold metadata",
			fmt.Sprintf("%s - Output: %s - Err: %s", err.Error(), outb.String(), errb.String()))
		return
	}
	metadata := &catalogc.Metadata{}
	err = protobuf.Decode(outb.Bytes(), metadata)
	if err != nil {
		task.CloseError(tef.Source, "failed to decode the scaffold metadata",
			err.Error())
		return
	}

	for key, vals := range values {
		// Special keys like '_method' start with _
		if len(key) > 0 && key[0] == '_' {
			continue
		}
		if len(vals) != 1 {
			task.CloseError(tef.Source, "unexpected number of values in form",
				fmt.Sprintf("for key '%s', got those values: '%v'", key, vals))
			return
		}
		val := vals[0]
		if len(val) == 0 {
			task.CloseError(tef.Source, "got an empty value",
				fmt.Sprintf("for key '%s', got this values: '%s'", key, val))
			return
		}
		metadata.UpdateOrSet(key, val)
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		task.CloseError(tef.Source, "failed to marshal metadata into JSON",
			err.Error())
		return
	}

	cmd = new(exec.Cmd)
	cmd = exec.Command("./pcadmin", "-c", conf.ConfigPath, "contract",
		"project", "invoke", "updateMetadata", "-i", p.InstanceID, "-bc", conf.BCPath,
		"-sign", conf.KeyID, "-metadataJSON", string(metadataJSON))

	task.AddInfo(tef.Source, "command created", fmt.Sprintf("%s", cmd.Args))
	outb.Reset()
	errb.Reset()
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	// We need to sleep because we just updated the status of projectc, which is
	// why we need to wait in order to have the right counter, otherwise we
	// could have an error of type "got counter=521, but need 522"
	time.Sleep(5 * time.Second)
	err = cmd.Run()
	if err != nil {
		task.CloseError(tef.Source, "failed to update the project instance",
			fmt.Sprintf("%s - Output: %s - Err: %s", err.Error(), outb.String(), errb.String()))
		return
	}
	cmdOut := outb.String()
	task.CloseOK(tef.Source, "project updated", "Output of the command: "+cmdOut)
}

// RequestUnlockEnclave talks to the enclave manager and asks it to boot an
// enclave.
func (p *Project) RequestUnlockEnclave(conf *Config) {
	tef := helpers.NewTaskEventFactory("ds manager")

	p.Status = ProjectStatusUnlockingEnclave
	p.StatusNotifier.Terminated = false
	p.StatusNotifier.UpdateStatus(ProjectStatusUnlockingEnclave)

	task := conf.TaskManager.NewTask("Request to unlock the enclave")

	// We use this client to listen to the events and update the project
	// status based on what we receive.
	client := task.Subscribe()

	task.AddInfo(tef.Source, "task created", "Hi from the RequestUnlockEnclave function")

	request := &Request{
		Description:    "Ask to unlock the enclave",
		Tasks:          []helpers.TaskI{task},
		StatusNotifier: helpers.NewStatusNotifier(),
		Status:         RequestStatusRunning,
	}
	p.AddRequest(request)

	// here we must check the cloud notifier instead of the task stream because
	// it will be the last on chain to say if the process succeeded. But we must
	// still check that the task stream ends well.
	go func() {
		for {
			select {
			case taskEl := <-client.TaskStream:
				if taskEl.Type == helpers.TypeCloseError {

					// Here we must check the error message and update the
					// FailedReasons of each attribute accordingly.
					_, latestDetails := p.GetLastestTaskMsg()
					splitMsg := strings.SplitN(latestDetails, "verification failed, here is why:\n", 2)
					log.LLvl1("here is the split", splitMsg)
					if len(splitMsg) == 2 {
						failedReasonsJSONStr := splitMsg[1]
						failedReasons := &catalogc.FailedReasons{}
						err := json.Unmarshal([]byte(failedReasonsJSONStr), failedReasons)
						if err != nil {
							log.Errorf("failed to unmarshal failed reasons: %s", err.Error())
						}
						p.updateFailedReasons(failedReasons, conf)
					}

					p.Status = ProjectStatusUnlockingEnclaveErrored
					p.StatusNotifier.UpdateStatusAndClose(
						ProjectStatusUnlockingEnclaveErrored)

					request.Status = RequestStatusErrored
					request.StatusNotifier.UpdateStatusAndClose(
						RequestStatusErrored)

				} else if taskEl.Type == helpers.TypeCloseOK {
					// nothing to do because there is still the cloud logs that
					// must be closed OK
				}
			default:
				select {
				case <-client.Done:
					return
				default:
				}
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	aliasName, bucketName, prefix := request.GetCloudAttributes(p.UID)
	cloudNotifier, err := helpers.NewCloudNotifier(aliasName, bucketName, prefix)
	if err != nil {
		log.Infof("failed to get the cloud notifier: %s", err.Error())
		task.CloseError(tef.Source, "failed to get the cloud notifier", err.Error())
		return
	}
	go func() {
		for {
			select {
			case taskEl := <-cloudNotifier.TaskEventCh:
				if taskEl.Type == helpers.TypeCloseError {

					p.Status = ProjectStatusUnlockingEnclaveErrored
					p.StatusNotifier.UpdateStatusAndClose(
						ProjectStatusUnlockingEnclaveErrored)

					request.Status = RequestStatusErrored
					request.StatusNotifier.UpdateStatusAndClose(
						RequestStatusErrored)

				} else if taskEl.Type == helpers.TypeCloseOK {

					p.Status = ProjectStatusUnlockingEnclaveDone
					p.StatusNotifier.UpdateStatusAndClose(
						ProjectStatusUnlockingEnclaveDone)

					request.Status = RequestStatusDone
					request.StatusNotifier.UpdateStatusAndClose(
						RequestStatusDone)

				}
			default:
				select {
				case <-cloudNotifier.Done:
					return
				default:
				}
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	urlStr := fmt.Sprintf("http://localhost:5000/eprojects/%s/unlock", p.InstanceID)

	task.AddInfo(tef.Source, "Sending a request to the enclave manager", " PUT "+urlStr)

	formData := url.Values{
		"_method":      {"put"},
		"requestIndex": {strconv.Itoa(request.Index)},
	}
	resp, err := http.PostForm(urlStr, formData)
	if err != nil {
		log.Infof("Failed to send request: %s", err.Error())
		task.CloseError(tef.Source, "Failed to send request", err.Error())
		return
	}
	defer resp.Body.Close()

	task.AddInfo(tef.Source, "Reading the status code", "Status: "+resp.Status)

	if resp.StatusCode != 200 {
		log.Infof("Got an unexpected response: %s", resp.Status)
		task.CloseError(tef.Source, "Got an unexpected response", "Status: "+resp.Status)
		return
	}

	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			log.Info("End of readloop")
			return
		}
		if err != nil {
			log.Infof("Response errored: %s", err.Error())
			task.CloseError(tef.Source, "Response errored", err.Error())
			return
		}

		if len(line) == 0 || string(line) == "\n" {
			continue
		}
		log.Infof("here is the task: %s", line)

		sections := bytes.SplitN(line, []byte(":"), 2)
		if len(sections) != 2 {
			log.Infof("Malformatted data: %s", line)
			task.CloseError(tef.Source, "Malformatted data", "Here is the line: "+string(line))
			return
		}
		field, value := string(sections[0]), sections[1]
		switch field {
		case "data":
			taskResponse, err := processStreamValue(value)
			if err != nil {
				log.Infof("Failed to read stream: %s", err.Error())
				task.CloseError(tef.Source, "Failed to read stream", err.Error())
				return
			}
			task.AddTaskEvent(*taskResponse)
		default:
			log.Infof("Unsupported field: %s", field)
			task.CloseError(tef.Source, "Unsupported field", "Here is the filed: "+field)
			return
		}
	}
}

func (p *Project) updateFailedReasons(failedReasons *catalogc.FailedReasons,
	conf *Config) error {

	// Get the project instance so we can extract the metadata, update it and
	// then send an update metadata request
	var projectContractData *projectc.ProjectData
	cmd := exec.Command("./pcadmin", "contract", "project", "get", "-i",
		p.InstanceID, "-bc", conf.BCPath, "-x")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to get the project instance: %s - "+
			"Output: %s - Err: %s", err.Error(), outb.String(), errb.String())
	}

	cmdOut := outb.Bytes()
	projectContractData = &projectc.ProjectData{}
	err = protobuf.Decode(cmdOut, projectContractData)
	if err != nil {
		return errors.New("failed to decode project innstance: " + err.Error())
	}
	for _, fr := range failedReasons.FailedReasons {
		if fr == nil {
			continue
		}
		attr, found := projectContractData.Metadata.GetAttribute(fr.AttributeID)
		if found {
			attr.AddFailedReason(fr.AttributeID, fr.Reason, fr.Dataset)
		}
	}

	// Now let's update the project instance with the metadata containing the
	// failed reasons
	jsonMetadata, err := json.Marshal(projectContractData.Metadata)
	if err != nil {
		return xerrors.Errorf("failed to marshal metadata into JSON: %v", err)
	}
	cmd = new(exec.Cmd)
	cmd = exec.Command("./pcadmin", "-c", conf.ConfigPath, "contract",
		"project", "invoke", "updateMetadata", "-bc", conf.BCPath, "-sign",
		conf.KeyID, "-metadataJSON", string(jsonMetadata), "-i", p.InstanceID)
	outb.Reset()
	errb.Reset()
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to update the metadata on the project: %s - "+
			"Output: %s - Err: %s", err.Error(), outb.String(), errb.String())
	}

	return nil
}

// RequestDeleteEnclave talks to the enclave manager and asks it to delete an
// enclave.
func (p *Project) RequestDeleteEnclave(conf *Config) {
	tef := helpers.NewTaskEventFactory("ds manager")

	p.Status = ProjectStatusDeletingEnclave
	p.StatusNotifier.Terminated = false
	p.StatusNotifier.UpdateStatus(ProjectStatusDeletingEnclave)

	task := conf.TaskManager.NewTask("Request to delete the enclave")

	// We use this client to listen to the events and update the project
	// status based on what we receive.
	client := task.Subscribe()

	task.AddInfo(tef.Source, "task created", "From the RequestDeleteEnclave function")

	request := &Request{
		Description:    "Ask to destroy the enclave",
		Tasks:          []helpers.TaskI{task},
		StatusNotifier: helpers.NewStatusNotifier(),
		Status:         RequestStatusRunning,
	}
	p.AddRequest(request)

	go func() {
		for {
			select {
			case taskEl := <-client.TaskStream:
				if taskEl.Type == helpers.TypeCloseError {

					p.Status = ProjectStatusDeletingEnclaveErrored
					p.StatusNotifier.UpdateStatusAndClose(
						ProjectStatusDeletingEnclaveErrored)

					request.Status = RequestStatusErrored
					request.StatusNotifier.UpdateStatusAndClose(
						RequestStatusErrored)

				} else if taskEl.Type == helpers.TypeCloseOK {

					p.Status = ProjectStatusDeletingEnclaveDone
					p.StatusNotifier.UpdateStatusAndClose(
						ProjectStatusDeletingEnclaveDone)

					request.Status = RequestStatusDone
					request.StatusNotifier.UpdateStatusAndClose(
						RequestStatusDone)
				}
			default:
				select {
				case <-client.Done:
					return
				default:
				}
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	formData := url.Values{
		"_method": {"delete"},
	}
	resp, err := http.PostForm("http://localhost:5000/eprojects/"+p.InstanceID, formData)
	if err != nil {
		log.Infof("Failed to send request: %s", err.Error())
		task.CloseError(tef.Source, "Failed to send request", err.Error())
		return
	}
	defer resp.Body.Close()

	task.AddInfo(tef.Source, "Reading the status code", "Status: "+resp.Status)

	if resp.StatusCode != 200 {
		log.Infof("Got an unexpected response: %s", resp.Status)
		task.CloseError(tef.Source, "Got an unexpected response", "Response: "+resp.Status)
		return
	}

	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			log.Info("End of readloop")
			return
		}
		if err != nil {
			log.Infof("Response errored: %s", err.Error())
			task.CloseError(tef.Source, "Response errored", err.Error())
			return
		}

		if len(line) == 0 || string(line) == "\n" {
			continue
		}
		log.Infof("here is the task: %s", line)

		sections := bytes.SplitN(line, []byte(":"), 2)
		if len(sections) != 2 {
			log.Infof("Malformatted data: %s", line)
			task.CloseError(tef.Source, "Malformatted data", "Line: "+string(line))
			return
		}
		field, value := string(sections[0]), sections[1]
		switch field {
		case "data":
			taskResponse, err := processStreamValue(value)
			if err != nil {
				log.Infof("Failed to read stream: %s", err.Error())
				task.CloseError(tef.Source, "Failed to read stream", err.Error())
				return
			}
			task.AddTaskEvent(*taskResponse)
		default:
			log.Infof("Unsupported field: %s", field)
			task.CloseError(tef.Source, "Unsupported field", "Field: "+field)
			return
		}
	}
}

// UpdateProjectcStatus updated the status of the project instance
func (p *Project) UpdateProjectcStatus(conf *Config, status string) error {
	if p.InstanceID == "" {
		return errors.New("instanceID field of project is empty")
	}

	var err error
	var outb, errb bytes.Buffer
	retry := 4

	for retry > 0 {
		outb.Reset()
		errb.Reset()
		// "A Cmd cannot be reused after calling its Run, Output or
		// CombinedOutput methods."
		cmd := exec.Command("./pcadmin", "-c", conf.ConfigPath, "contract",
			"project", "invoke", "updateStatus", "-bc", conf.BCPath, "-sign",
			conf.KeyID, "-status", status, "-i", p.InstanceID)
		cmd.Stdout = &outb
		cmd.Stderr = &errb

		err = cmd.Run()
		if err == nil {
			break
		}
		retry--
		log.Warnf("failed to update the projectc status, %d try left, "+
			"sleeping for 7s...: %s", retry, err.Error())
		time.Sleep(7 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to update the project instance status after 4 try, "+
			"failed to run the command: %s - "+
			"Output: %s - Err: %s", err.Error(), outb.String(), errb.String())
	}

	return nil
}
