package controllers

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"time"

	xhelpers "github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/dedis/odyssey/enclavem/app/helpers"
	"github.com/dedis/odyssey/enclavem/app/models"
	"github.com/dedis/odyssey/projectc"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/protobuf"
)

// EProjectsIndexHandler ...
func EProjectsIndexHandler(store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			eProjectsIndexGet(w, r, store)
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				xhelpers.RedirectWithErrorFlash("/", "failed to read form", w, r, store)
				return
			}
			switch r.PostFormValue("_method") {
			case "delete":
				eProjectsIndexDelete(w, r, store)
			default:
				xhelpers.RedirectWithErrorFlash("/", "only DELETE allowed", w, r, store)
			}
		}
	}
}

// EProjectsShowHandler ...
func EProjectsShowHandler(store *sessions.CookieStore, conf *models.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			eProjectsShowGet(w, r, store)
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				xhelpers.RedirectWithErrorFlash(r.URL.String(), "failed to read form", w, r, store)
				return
			}
			switch r.PostFormValue("_method") {
			case "delete":
				eProjectsShowDelete(w, r, store, conf)
			default:
				xhelpers.RedirectWithErrorFlash(r.URL.String(), "only DELETE allowed", w, r, store)
			}
		}
	}
}

// EProjectsShowUnlockHandler ...
func EProjectsShowUnlockHandler(store *sessions.CookieStore, conf *models.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				xhelpers.RedirectWithErrorFlash(r.URL.String(), "failed to read form", w, r, store)
				return
			}
			switch r.PostFormValue("_method") {
			case "put":
				eProjectsShowUnlockPost(w, r, store, conf)
			default:
				xhelpers.RedirectWithErrorFlash(r.URL.String(), "only PUT allowed", w, r, store)
			}
		}
	}
}

func eProjectsIndexGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore) {

	t, err := template.ParseFiles("views/layout.gohtml", "views/eprojects/index.gohtml")
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	type viewData struct {
		Title     string
		EProjects []*models.EProject
		Flash     []xhelpers.Flash
	}

	flashes, err := xhelpers.ExtractFlash(w, r, store)
	if err != nil {
		fmt.Printf("Failed to get flash: %s\n", err.Error())
	}

	projectSlice := []*models.EProject{}
	for _, value := range models.EProjectList {
		projectSlice = append(projectSlice, value)
	}

	p := &viewData{
		Title:     "List of projects",
		Flash:     flashes,
		EProjects: projectSlice,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		fmt.Printf("Error while executing template: %s\n", err.Error())
	}
}

func eProjectsIndexDelete(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore) {

	for k, eproject := range models.EProjectList {
		token, err := helpers.GetToken(w)
		if err != nil {
			xhelpers.RedirectWithErrorFlash("/eprojects", "failed to get authentication token: "+err.Error(), w, r, store)
			return
		}

		req, err := http.NewRequest("DELETE", eproject.EnclaveHref, nil)
		if err != nil {
			xhelpers.RedirectWithErrorFlash("/eprojects", "failed to create request: "+err.Error(), w, r, store)
			return
		}
		req.Header.Set("x-vcloud-authorization", token)
		req.Header.Set("Accept", "application/*+xml;version=31.0")

		client := &http.Client{}
		log.LLvlf1("sending POST request to get vApp: %s", eproject.EnclaveHref)
		resp, err := client.Do(req)
		if err != nil {
			xhelpers.RedirectWithErrorFlash("/eprojects", "failed to send request: "+err.Error(), w, r, store)
			return
		}

		if resp.StatusCode != 202 {
			bodyBuf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				xhelpers.RedirectWithErrorFlash("/eprojects", fmt.Sprintf(
					"got an unexpected status code %s and failed to read body:%s ",
					resp.Status, err.Error()), w, r, store)
				return
			}
			var vcloudError models.VcloudError
			err = xml.Unmarshal(bodyBuf, &vcloudError)
			if err != nil {
				xhelpers.RedirectWithErrorFlash("/eprojects", fmt.Sprintf(
					"got an unexpected status code %s and failed to unmashall body:%s ",
					resp.Status, err.Error()), w, r, store)
				return
			}
			xhelpers.RedirectWithErrorFlash("/eprojects", fmt.Sprintf(
				"expected 200, got %s with the following message: %s",
				resp.Status, vcloudError.Message), w, r, store)
			return
		}

		delete(models.EProjectList, k)
	}

	xhelpers.RedirectWithInfoFlash("/eprojects", "eprojects and enclaves deleted", w, r, store)
}

func eProjectsShowGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore) {

	params := mux.Vars(r)
	id := params["instID"]
	if id == "" {
		xhelpers.RedirectWithErrorFlash("/", "instance ID not found in URL", w, r, store)
		return
	}

	eproject, ok := models.EProjectList[id]
	if !ok || eproject == nil {
		xhelpers.RedirectWithErrorFlash("/", "eproject not found", w, r, store)
		return
	}

	t, err := template.ParseFiles("views/layout.gohtml", "views/eprojects/show.gohtml")
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	type viewData struct {
		Title    string
		EProject *models.EProject
		Flash    []xhelpers.Flash
	}

	flashes, err := xhelpers.ExtractFlash(w, r, store)
	if err != nil {
		fmt.Printf("Failed to get flash: %s\n", err.Error())
	}

	p := &viewData{
		Title:    "List of projects",
		Flash:    flashes,
		EProject: eproject,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		fmt.Printf("Error while executing template: %s\n", err.Error())
	}

}

// Called from the data scientist manager
func eProjectsShowDelete(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Error("streaming not supported")
		return
	}

	tef := xhelpers.NewTaskEventFFactory("enclave manager", flusher, w)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "false")

	tef.FlushTaskEventInfo("hi from the enclave manager",
		"let's destroy this VM - from eProjectsShowDelete")

	params := mux.Vars(r)
	id := params["instID"]
	if id == "" {
		tef.FlushTaskEventCloseError("instance ID not found in URL", id)
		return
	}

	eproject, ok := models.EProjectList[id]
	if !ok || eproject == nil {
		tef.FlushTaskEventCloseError("eproject not found", id)
		return
	}
	tef.FlushTaskEventInfo("EProject found", fmt.Sprintf("%v", eproject))

	// This should be called before exiting the function when an error occurs,
	// it takes care of updating the project instance status and if the update
	// doesn't work it updates the error message to tell the user what
	// happened.
	handleError := func(msg string, details string, args ...interface{}) {
		err2 := models.UpdateProjectcStatus(conf, "deletedErrored", eproject.InstanceID)
		if len(args) == 0 {
			tef.XFlushTaskEventCloseError(err2, msg, details)
		} else {
			tef.XFlushTaskEventCloseErrorf(err2, msg, details, args...)
		}
		return
	}

	tef.FlushTaskEventInfo("updating the contract status",
		"setting the status to 'deleting' in the project's instance")
	err := models.UpdateProjectcStatus(conf, "deleting", eproject.InstanceID)
	if err != nil {
		handleError("failed to update the contract status", err.Error())
		return
	}

	token, err := helpers.GetToken(w)
	if err != nil {
		handleError("failed to get authentication token", err.Error())
		return
	}

	// ------------------------------------------------------------------------
	// GET THE VAPP

	url := eproject.EnclaveHref
	resp, err := getVapp(url, token)
	defer resp.Body.Close()
	if err != nil {
		handleError("failed to get vApp", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}

	if resp.StatusCode != 200 {
		bodyBuf, err := ioutil.ReadAll(resp.Body)
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		if err != nil {
			handleError("unexpected status code", "got an "+
				"unexpected status code %s and failed to read body:%s ",
				resp.Status, err.Error())
			return
		}
		var vcloudError models.VcloudError
		err = xml.Unmarshal(bodyBuf, &vcloudError)
		if err != nil {
			handleError("unexpected status code", "got an "+
				"unexpected status code %s and failed to unmashall body:%s ",
				resp.Status, err.Error())
			return
		}
		handleError("unexpected status code", "expected 200, "+
			"got %s with the following message: %s", resp.Status, vcloudError.Message)
		return
	}

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError("failed to read body", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}
	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	var vapp models.VApp
	err = xml.Unmarshal(bodyBuf, &vapp)
	if err != nil {
		handleError("failed to unmarshal body", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}

	// ------------------------------------------------------------------------
	// POWEROFF THE VAPP
	var taskURL string
	var task models.Task
	var client *http.Client
	var req *http.Request
	var undeployArg []byte

	// 4 = Powered on
	// 8 = Powered off
	// 3 = Suspended
	if vapp.Status == "8" {
		tef.FlushTaskEventInfo("enclave already powered off, skipping",
			"we skip the undeploy part since the enclave has status 8")
		goto removeVapp
	}

	log.Lvl1("powerOff (undeploy) the newly created vApp")
	tef.FlushTaskEventInfo("undeloy the vApp", "force shut down (undeploy) the newly created vApp")
	// Here we powerOff instead of shutdown the vApp because we don't care
	// anymore about the content of the vApp.
	undeployArg = []byte(`<?xml version="1.0" encoding="UTF-8"?>
	<vcloud:UndeployVAppParams
		xmlns:vcloud="http://www.vmware.com/vcloud/v1.5">
		<vcloud:UndeployPowerAction>powerOff</vcloud:UndeployPowerAction> 
	</vcloud:UndeployVAppParams>`)

	url = fmt.Sprintf("%s/action/undeploy", eproject.EnclaveHref)
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(undeployArg))
	if err != nil {
		handleError("failed to create request", err.Error())
		return
	}
	req.Header.Set("x-vcloud-authorization", token)
	req.Header.Set("Accept", "application/*+xml;version=31.0")
	req.Header.Set("Content-Type", "application/vnd.vmware.vcloud.undeployVAppParams+xml")

	client = &http.Client{}
	log.Lvlf1("sending the POST request to poweroff (undeploy) the vApp: %s", url)
	tef.FlushTaskEventInfo("send the request", "sending the POST request to poweroff (undeploy) the vApp")

	resp, err = client.Do(req)
	if err != nil {
		handleError("failed to send request", err.Error())
		return
	}
	defer resp.Body.Close()

	log.Lvlf1("got back this status code: %s", resp.Status)
	tef.FlushTaskEventInfo("read the status code", resp.Status)

	if resp.StatusCode != 202 {
		bodyBuf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			handleError("unexpected status code", "got an "+
				"unexpected status code %s and failed to read body:%s ",
				resp.Status, err.Error())
			return
		}
		var vcloudError models.VcloudError
		err = xml.Unmarshal(bodyBuf, &vcloudError)
		if err != nil {
			handleError("unexpected status code", "got an "+
				"unexpected status code %s and failed to unmashall body:%s ",
				resp.Status, err.Error())
			return
		}
		handleError("unexpected status code", "expected 202, "+
			"got %s with the following message: %s", resp.Status, vcloudError.Message)
		return
	}

	bodyBuf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError("failed to read body", err.Error())
		return
	}

	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	err = xml.Unmarshal(bodyBuf, &task)
	if err != nil {
		handleError("failed to unmarshal body", err.Error())
		return
	}

	// ------------------------------------------------------------------------
	// WAIT FOR THE SHUTDOWN TASK TO END

	taskURL = task.Href
	log.Lvlf1("waiting for the POST shutdown task to succeed: %s", taskURL)
	tef.FlushTaskEventInfo("waiting for the POST shutdown task to succeed", taskURL)

	err = pollTask(tef, taskURL, token, 10, time.Duration(time.Second*5), 1.2)
	if err != nil {
		handleError("failed to poll the shutdown task", err.Error())
		return
	}

	tef.FlushTaskEventImportantInfo("Vm is now shut down", "")

	// ------------------------------------------------------------------------
	// DELETE THE VAPP
removeVapp:
	req, err = http.NewRequest("DELETE", eproject.EnclaveHref, nil)
	if err != nil {
		handleError("failed to create request", err.Error())
		return
	}
	req.Header.Set("x-vcloud-authorization", token)
	req.Header.Set("Accept", "application/*+xml;version=31.0")

	client = &http.Client{}
	log.LLvlf1("sending DELETE request to vApp: %s", eproject.EnclaveHref)
	tef.FlushTaskEventInfo("sending DELETE request to delete the vApp", eproject.EnclaveHref)
	resp, err = client.Do(req)
	if err != nil {
		handleError("failed to send request", err.Error())
		return
	}

	if resp.StatusCode != 202 {
		bodyBuf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			handleError("unexpected status code", "got an "+
				"unexpected status code %s and failed to read body:%s ",
				resp.Status, err.Error())
			return
		}
		var vcloudError models.VcloudError
		err = xml.Unmarshal(bodyBuf, &vcloudError)
		if err != nil {
			handleError("unexpected status code", "got an "+
				"unexpected status code %s and failed to unmashall body:%s ",
				resp.Status, err.Error())
			return
		}
		handleError("unexpected status code", "expected 202, "+
			"got %s with the following message: %s", resp.Status, vcloudError.Message)
		return
	}

	bodyBuf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError("failed to read body", err.Error())
		return
	}

	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	err = xml.Unmarshal(bodyBuf, &task)
	if err != nil {
		handleError("failed to unmashal body", err.Error())
		return
	}

	// ------------------------------------------------------------------------
	// WAIT FOR THE DELETE TASK TO END

	taskURL = task.Href
	log.Lvlf1("waiting for the delete task to succeed: %s", taskURL)
	tef.FlushTaskEventInfo("waiting for the delete task to succeed", taskURL)

	err = pollTask(tef, taskURL, token, 10, time.Duration(time.Second*5), 1.2)
	if err != nil {
		handleError("failed to poll the delete task", err.Error())
		return
	}

	tef.FlushTaskEventImportantInfo("vM deleted", "")

	tef.FlushTaskEventInfo("updating the contract status",
		"setting the status to 'deletedOK' in the project's instance")
	err = models.UpdateProjectcStatus(conf, "deletedOK", eproject.InstanceID)
	if err != nil {
		handleError("failed to update the contract status", err.Error())
		return
	}

	delete(models.EProjectList, id)

	err = models.UpdateProjectcStatus(conf, "deletedOK", eproject.InstanceID)
	if err != nil {
		handleError("failed to update the contract status",
			"failed to update the status to 'deletedOK", err)
		return
	}

	tef.FlushTaskEventCloseOK("vApp destroyed ", eproject.EnclaveName)

}

func eProjectsShowUnlockPost(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {
	// We assume that r.ParseForm() has already been called

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Error("streaming not supported")
		return
	}

	tef := xhelpers.NewTaskEventFFactory("enclave manager", flusher, w)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "false")

	tef.FlushTaskEventInfo("hi from the enclave manager", "let's try to unlock this VM - from eProjectsShowUnlockPost")

	params := mux.Vars(r)
	id := params["instID"]
	if id == "" {
		tef.FlushTaskEventCloseError("failed to get the instance ID", id)
		return
	}

	eproject, ok := models.EProjectList[id]
	if !ok || eproject == nil {
		tef.FlushTaskEventCloseError("EProject not found", id)
		return
	}
	tef.FlushTaskEventInfo("EProject found", fmt.Sprintf("%v", eproject))

	// This should be called before exiting the function when an error occurs,
	// it takes care of updating the project instance status and if the update
	// doesn't work it updates the error message to tell the user what
	// happened.
	handleError := func(msg string, details string, args ...interface{}) {
		err2 := models.UpdateProjectcStatus(conf, "unlockedErrored", eproject.InstanceID)
		if len(args) == 0 {
			tef.XFlushTaskEventCloseError(err2, msg, details)
		} else {
			tef.XFlushTaskEventCloseErrorf(err2, msg, details, args...)
		}
		return
	}

	tef.FlushTaskEventInfo("updating the contract status",
		"setting the status to 'unlocking' in the project's instance")
	err := models.UpdateProjectcStatus(conf, "unlocking", eproject.InstanceID)
	if err != nil {
		handleError("failed to update the contract status", err.Error())
		return
	}

	eproject.Status = models.EProjectStatusUnlockingEnclave

	log.Lvlf1("We got this post form: %v", r.PostForm)
	tef.FlushTaskEventInfo("getting the request index", "now let's try to get the request index from the POST form")

	requestIndex := r.PostForm.Get("requestIndex")
	if requestIndex == "" {
		handleError("Got en empty requestIndex", requestIndex)
		return
	}
	tef.FlushTaskEventInfo("request index found", requestIndex)
	tef.FlushTaskEventInfo("getting the project", "running 'pcadmin contract project get")
	cmd := exec.Command("./pcadmin", "contract", "project", "get", "-i", eproject.InstanceID, "-bc", conf.BCPath, "-x")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		handleError("failed to get the project instance",
			"failed to get the project instance: %s - Output: %s - Err: %s",
			err.Error(), outb.String(), errb.String())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}
	cmdOut := outb.Bytes()
	projectContractData := &projectc.ProjectData{}
	err = protobuf.Decode(cmdOut, projectContractData)
	if err != nil {
		handleError("failed to decode project innstance", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}

	tef.FlushTaskEventInfof("response decoded", "we got the project instance: %v", projectContractData)

	enclavePubKey, err := eproject.ParseKey()
	if err != nil {
		handleError("failed to parse the public key", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}

	readInstIDSlice := make([]string, 0)
	// The write instance ids correspond to the "Datasets stored in the project
	// contract".
	writeInstIDSlice := make([]string, 0)

	for _, instID := range projectContractData.Datasets {
		instIDStr := instID.String()
		writeInstIDSlice = append(writeInstIDSlice, instIDStr)
		tef.FlushTaskEventInfof("sleeping", "sleeping 10 sec before talking to cothority...")
		time.Sleep(time.Second * 10)
		cmd := exec.Command("./csadmin", "-c", conf.ConfigPath, "contract",
			"read", "spawn", "-i", instIDStr, "-bc", conf.BCPath,
			"-pid", eproject.InstanceID, "-s", conf.KeyID,
			"--key", enclavePubKey, "-x")
		tef.FlushTaskEventInfof("trying to spawn a read instance", fmt.Sprintf("%v", cmd.Args))
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err := cmd.Run()
		if err != nil {
			handleError("failed to run the spawn cmd", "failed to spawn the read "+
				"instance: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String())
			eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
			return
		}

		readInstID := outb.String()
		instIDPattern, err := regexp.Compile("^[0-9a-f]{64}$")
		if err != nil {
			handleError("failed to build instID regex", err.Error())
			eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
			return
		}

		if !instIDPattern.MatchString(readInstID) {
			handleError("wrong instance id", "the content "+
				"of read instance ID is unexpected, found '%s'", readInstID)
			eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
			return
		}

		readInstIDSlice = append(readInstIDSlice, readInstID)
		tef.FlushTaskEventInfof("dataset validated", "ok for dataset '%s', "+
			"got this read instance id: %s", instIDStr, readInstID)
	}

	eproject.ReadInstIDs = readInstIDSlice
	eproject.WriteInstIDs = writeInstIDSlice

	tef.FlushTaskEventImportantInfo("All checks passed", "we can download "+
		"the datasets on the enclave and unlock it")

	// ------------------------------------------------------------------------
	// GET THE VAPP

	token, err := helpers.GetToken(w)
	if err != nil {
		handleError("failed to get token", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}

	url := eproject.EnclaveHref
	resp, err := getVapp(url, token)
	defer resp.Body.Close()
	if err != nil {
		handleError("failed to get vApp", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}

	if resp.StatusCode != 200 {
		bodyBuf, err := ioutil.ReadAll(resp.Body)
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		if err != nil {
			handleError("unexpected status code", "got an "+
				"unexpected status code %s and failed to read body:%s ",
				resp.Status, err.Error())
			return
		}
		var vcloudError models.VcloudError
		err = xml.Unmarshal(bodyBuf, &vcloudError)
		if err != nil {
			handleError("unexpected status code", "got an "+
				"unexpected status code %s and failed to unmashall body:%s ",
				resp.Status, err.Error())
			return
		}
		handleError("unexpected status code", "expected 200, "+
			"got %s with the following message: %s", resp.Status, vcloudError.Message)
		return
	}

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError("failed to read body", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}
	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	var vapp models.VApp
	err = xml.Unmarshal(bodyBuf, &vapp)
	if err != nil {
		handleError("failed to unmarshal body", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}

	// ------------------------------------------------------------------------
	// SET THE PROPERTIES ON THE VM (LIST OF READ INSTANCE IDS)

	vmLink := vapp.Children.VM.Href

	log.Lvl1("setting property on the VM (the read instIDs) with link " + vmLink)
	tef.FlushTaskEventInfo("setting property on the VM (the read instIDs)", vmLink)

	propertyParam := struct {
		ReadInstIDs  string
		WriteInstIDs string
		// The bucket where to store the cloud stuff at dedis/{bucket}
		Endpoint string
		// The public key of the data scientist manager
		PubKey string
		// Used to publish the cloud logs at "logs/{request index}"
		RequestIndex string
		// The project instance ID
		ProjectIID string
		// Darc ID needed to set the ready signal on the project instance
		DarcID string
	}{
		strings.Join(eproject.ReadInstIDs, ","),
		strings.Join(eproject.WriteInstIDs, ","),
		eproject.CloudEndpoint,
		projectContractData.AccessPubKey,
		requestIndex,
		eproject.InstanceID,
		conf.DarcID,
	}
	// When updating we have to include every properties we want to keep. The
	// update does not keep the older properties.
	t, err := template.New("propertyArg").Parse(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
	<ns6:ProductSectionList xmlns="http://www.vmware.com/vcloud/versions"
		xmlns:ns2="http://schemas.dmtf.org/ovf/envelope/1"
		xmlns:ns3="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData"
		xmlns:ns4="http://schemas.dmtf.org/wbem/wscim/1/common"
		xmlns:ns5="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData"
		xmlns:ns6="http://www.vmware.com/vcloud/v1.5"
		xmlns:ns7="http://www.vmware.com/schema/ovf"
		xmlns:ns8="http://schemas.dmtf.org/ovf/environment/1"
		xmlns:ns9="http://www.vmware.com/vcloud/extension/v1.5">
		<ns2:ProductSection>
			<ns2:Info ns2:msgid="productSection"> Product section info </ns2:Info>
			<ns2:Property ns2:userConfigurable="false" ns2:type="string" ns2:password="false" ns2:key="ready_to_download" ns2:value="go"/>
			<ns2:Property ns2:userConfigurable="false" ns2:type="string" ns2:password="false" ns2:key="read_instance_ids" ns2:value="{{ .ReadInstIDs}}"/>
			<ns2:Property ns2:userConfigurable="false" ns2:type="string" ns2:password="false" ns2:key="write_instance_ids" ns2:value="{{ .WriteInstIDs}}"/>
			<ns2:Property ns2:userConfigurable="false" ns2:type="string" ns2:password="false" ns2:key="vm_endpoint" ns2:value="{{ .Endpoint}}"/>
			<ns2:Property ns2:userConfigurable="false" ns2:type="string" ns2:password="false" ns2:key="authorized_key" ns2:value="{{ .PubKey}}"/>
			<ns2:Property ns2:userConfigurable="false" ns2:type="string" ns2:password="false" ns2:key="request_index" ns2:value="{{ .RequestIndex}}"/>
			<ns2:Property ns2:userConfigurable="false" ns2:type="string" ns2:password="false" ns2:key="project_instance_id" ns2:value="{{ .ProjectIID}}"/>
			<ns2:Property ns2:userConfigurable="false" ns2:type="string" ns2:password="false" ns2:key="darc_id" ns2:value="{{ .DarcID}}"/>
		</ns2:ProductSection>
	</ns6:ProductSectionList>
	`)
	if err != nil {
		handleError("failed to create template", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		err = t.Execute(pw, propertyParam)
		if err != nil {
			handleError("failed to execute template", err.Error())
			eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
			return
		}
	}()

	url = fmt.Sprintf("%s/productSections", vmLink)
	req, err := http.NewRequest("PUT", url, pr)
	if err != nil {
		handleError("failed to create request", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}
	req.Header.Set("x-vcloud-authorization", token)
	req.Header.Set("Accept", "application/*+xml;version=31.0")
	req.Header.Set("Content-Type", "application/vnd.vmware.vcloud.productSections+xml")

	client := &http.Client{}
	log.LLvlf1("sending PUT request to update properties with url: %s", url)
	tef.FlushTaskEventInfo("sending PUT request to update properties", url)

	resp, err = client.Do(req)
	if err != nil {
		handleError("failed to send request", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}
	defer resp.Body.Close()

	log.Lvlf1("got back this status code: %s", resp.Status)
	tef.FlushTaskEventInfo("got back this status code", resp.Status)

	if resp.StatusCode != 202 {
		bodyBuf, err := ioutil.ReadAll(resp.Body)
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		if err != nil {
			handleError("unexpected status code", "got an "+
				"unexpected status code %s and failed to read body:%s ",
				resp.Status, err.Error())
			return
		}
		var vcloudError models.VcloudError
		err = xml.Unmarshal(bodyBuf, &vcloudError)
		if err != nil {
			handleError("unexpected status code", "got an "+
				"unexpected status code %s and failed to unmashall body:%s ",
				resp.Status, err.Error())
			return
		}
		handleError("unexpected status code", "expected 202, "+
			"got %s with the following message: %s", resp.Status, vcloudError.Message)
		return
	}

	bodyBuf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError("failed to read body", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}

	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	var task models.Task
	err = xml.Unmarshal(bodyBuf, &task)
	if err != nil {
		handleError("failed to unmashal body", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}

	// ------------------------------------------------------------------------
	// WAIT FOR NEW PROPERTY TASK TO END

	taskURL := task.Href
	log.Lvlf1("waiting for the PUT properties task to succeed: %s", taskURL)
	tef.FlushTaskEventInfo("waiting for the PUT properties task to succeed", taskURL)

	err = pollTask(tef, taskURL, token, 10, time.Duration(time.Second*5), 1.2)
	if err != nil {
		handleError("failed to poll the set property task", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}

	tef.FlushTaskEventImportantInfo("Property (cloud endpoint to use) on Vm set", taskURL)

	// ------------------------------------------------------------------------
	// POWER ON THE NEWLY CREATED VAPP

	log.Lvl1("powering on the newly created vApp")
	tef.FlushTaskEventInfo("powering on the vApp", vapp.Href)

	url = fmt.Sprintf("%s/power/action/powerOn", vapp.Href)
	req, err = http.NewRequest("POST", url, nil)
	if err != nil {
		handleError("failed to create request", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}
	req.Header.Set("x-vcloud-authorization", token)
	req.Header.Set("Accept", "application/*+xml;version=31.0")

	client = &http.Client{}
	log.Lvlf1("sending the POST request to power on the vApp: %s", url)
	tef.FlushTaskEventInfo("sending the POST request to power on the vApp", url)

	resp, err = client.Do(req)
	if err != nil {
		handleError("failed to send request", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}
	defer resp.Body.Close()

	log.Lvlf1("got back this status code: %s", resp.Status)
	tef.FlushTaskEventInfo("reading status code", resp.Status)

	if resp.StatusCode != 202 {
		bodyBuf, err := ioutil.ReadAll(resp.Body)
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		if err != nil {
			handleError("unexpected status code", "got an "+
				"unexpected status code %s and failed to read body:%s ",
				resp.Status, err.Error())
			return
		}
		var vcloudError models.VcloudError
		err = xml.Unmarshal(bodyBuf, &vcloudError)
		if err != nil {
			handleError("unexpected status code", "got an "+
				"unexpected status code %s and failed to unmashall body:%s ",
				resp.Status, err.Error())
			return
		}
		handleError("unexpected status code", "expected 202, "+
			"got %s with the following message: %s", resp.Status, vcloudError.Message)
		return
	}

	bodyBuf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError("failed to read body", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}

	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	err = xml.Unmarshal(bodyBuf, &task)
	if err != nil {
		handleError("failed to unmashal body", err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}

	// ------------------------------------------------------------------------
	// WAIT FOR THE POWER ON TASK TO END

	taskURL = task.Href
	log.Lvlf1("waiting for the POST power on task to succeed: %s", taskURL)
	tef.FlushTaskEventInfo("waiting for the POST power on task to succeed", taskURL)

	err = pollTask(tef, taskURL, token, 10, time.Duration(time.Second*5), 1.3)
	if err != nil {
		handleError("failed to poll the power on task. "+
			"Maybe the IP adress %s is already used?\nError: %s",
			eproject.IPAddr.String(), err.Error())
		eproject.Status = models.EProjectStatusUnlockingEnclaveErrored
		return
	}

	tef.FlushTaskEventImportantInfo("Vm powered on", "")

	tef.FlushTaskEventInfo("updating the contract status",
		"setting the status to 'unlockedOK' in the project's instance")
	err = models.UpdateProjectcStatus(conf, "unlockedOK", eproject.InstanceID)
	if err != nil {
		handleError("failed to update the contract status", err.Error())
		return
	}

	tef.FlushTaskEventCloseOK("enclave unlocked", eproject.EnclaveName)
}
