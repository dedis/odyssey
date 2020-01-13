package controllers

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"time"

	xhelpers "github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/dedis/odyssey/enclavem/app/helpers"
	"github.com/dedis/odyssey/enclavem/app/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/minio/minio-go/v6"
	"go.dedis.ch/onet/v3/log"
)

// VappsIndexHandler points to:
// GET /Vapps
// POST /Vapps
func VappsIndexHandler(gs *sessions.CookieStore, conf *models.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			VappsIndexGet(w, r, gs)
		case http.MethodPost:
			VappsIndexPost(w, r, gs, conf)
		}
	}
}

// VappsShowHandler points to:
// GET /Vapps/{id}
func VappsShowHandler(gs *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			VappsShowGet(w, r, gs)
		}
	}
}

// VappsIndexGet retrives the list of vApps and sends its json representation
func VappsIndexGet(w http.ResponseWriter, r *http.Request, gs *sessions.CookieStore) {

	token, err := helpers.GetToken(w)
	if err != nil {
		helpers.SendRequestError(err, w)
		return
	}

	url := fmt.Sprintf("https://%s/api/vApps/query", helpers.VcdHost)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.SendRequestError(errors.New("failed to build request:"+err.Error()), w)
		return
	}
	req.Header.Set("x-vcloud-authorization", token)
	req.Header.Set("Accept", "application/*;version=31.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		helpers.SendRequestError(errors.New("failed to send request: "+err.Error()), w)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		helpers.SendRequestError(helpers.NewNotOkError(resp), w)
		return
	}

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		helpers.SendRequestError(errors.New("failed to read body: "+err.Error()), w)
		return
	}
	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	var Vapps models.QueryResultRecords
	err = xml.Unmarshal(bodyBuf, &Vapps)
	if err != nil {
		helpers.SendRequestError(errors.New("failed to unmarshal body: "+err.Error()), w)
		return
	}

	js, err := json.MarshalIndent(Vapps, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// VappsIndexPost creates a new vApp based on a template. It sets the correct
// properties and network configurations and checks if the VM of the vApp has
// written its newly created public key on the cloud endpoint. If everything
// goes well the vApp is power off.
// TODO: if an error occurs we should ensure that the created vApp is destroyed
func VappsIndexPost(w http.ResponseWriter, r *http.Request, gs *sessions.CookieStore, conf *models.Config) {
	// Apparently, we must do this at the begining, after what the reader is
	// closed and we get a "http: invalid Read on closed Body"
	err := r.ParseForm()

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

	tef.FlushTaskEventInfo("hi from the enclave manager", "let's prpare this VM - from VappsIndexPost")

	var taskURL string
	var project *models.EProject
	var vapp models.VApp
	var bodyBuf []byte
	var err2 error

	tef.FlushTaskEventInfo("parse arguments", "Trying to parse the POST arguments")

	if err != nil {
		tef.FlushTaskEventCloseError("failed to parse POST argument", err.Error())
		return
	}
	log.Lvlf1("We got this post form: %v", r.PostForm)

	instIDR, err := regexp.Compile("^[0-9a-f]{64}$")
	if err != nil {
		tef.FlushTaskEventCloseError("failed to compile regex: ", err.Error())
		return
	}

	projectInstIDs, ok := r.PostForm["projectInstID"]
	if !ok || len(projectInstIDs) != 1 {
		tef.FlushTaskEventCloseErrorf("failed to parse the instance IDs",
			"the form key 'datasetIDs' is empty or contains more than one "+
				"element: %s", projectInstIDs)
		return
	}

	projectInstID := projectInstIDs[0]
	ok = instIDR.MatchString(projectInstID)
	if !ok {
		tef.FlushTaskEventCloseError("got a wong project instance ID", projectInstID)
		return
	}

	log.Info("project instance ID: " + projectInstID)
	tef.FlushTaskEventInfo("successfully got the project instance ID", projectInstID)

	// This should be called before exiting the function when an error occurs,
	// it takes care of updating the project instance status and if the update
	// doesn't work it updates the error message to tell the user what
	// happened.
	handleError := func(msg string, details string, args ...interface{}) {
		err2 = models.UpdateProjectcStatus(conf, "preparedErrored", projectInstID)
		if len(args) == 0 {
			tef.XFlushTaskEventCloseError(err2, msg, details)
		} else {
			tef.XFlushTaskEventCloseErrorf(err2, msg, details, args...)
		}
		return
	}

	tef.FlushTaskEventInfo("updating the contract status",
		"setting 'preparing' on the project's instance status ")
	err = models.UpdateProjectcStatus(conf, "preparing", projectInstID)
	if err != nil {
		handleError("failed to update the contract status", err.Error())
		return
	}

	tef.FlushTaskEventInfo("getting argument", "now let's try to get the project UID from the POST form")
	projectUID := r.PostForm.Get("projectUID")
	if projectUID == "" {
		handleError("Got en empty project UID", projectUID)
		return
	}
	tef.FlushTaskEventInfo("got the project UID", fmt.Sprintf("alright, we got this project UID: %s."+
		"This UID will be used as the cloud endpoint and the enclave UID", projectUID))

	tef.FlushTaskEventInfo("getting the request index", "now let's try to get the request index from the POST form")
	requestIndex := r.PostForm.Get("requestIndex")
	if requestIndex == "" {
		handleError("Got en empty requestIndex", requestIndex)
		return
	}
	tef.FlushTaskEventInfo("got the index request", fmt.Sprintf(
		"alright, we got this request index: %s. This request index will be "+
			"used by the enclave to send logs at /{projectUID}/logs/{requestIndex}/{timestamp}", requestIndex))

	tef.FlushTaskEventInfo("trying to authenticate", "trying to get an authentication token for vCloud")
	token, err := helpers.GetToken(w)
	if err != nil {
		handleError("failed to get authentication token", err.Error())
		return
	}
	tef.FlushTaskEventInfo("authentication OK", "we got the authentication token")

	tef.FlushTaskEventInfo("getting minio client", "creating the minio client based on the ENV variables")
	minioClient, err := xhelpers.GetMinioClient()
	if err != nil {
		handleError("failed to get the minion client", err.Error())
		return
	}

	tef.FlushTaskEventInfof("creating the bucket", "now we are creating the bucket '%s'", projectUID)
	found, err := minioClient.BucketExists(projectUID)

	if err != nil {
		handleError("failed to check the bucket",
			"failed to check if bucket '%s' exist: %s", projectUID, err.Error())
		return
	}
	if !found {
		err = minioClient.MakeBucket(projectUID, "us-east-1")
		if err != nil {
			handleError("failed to make the bucket", err.Error())
			return
		}
		tef.FlushTaskEventInfo("bucket created", "successfully created the bucket!")
	} else {
		tef.FlushTaskEventInfo("bucket not created", "bucket not created because it is already created")
	}

	// ------------------------------------------------------------------------
	// CREATE NEW APP
	log.LLvl1("creating new app via POST")
	tef.FlushTaskEventInfo("creating vApp", "creating new app via POST")

	nameBuf := make([]byte, 8)
	_, err = rand.Read(nameBuf)
	if err != nil {
		handleError("failed to create random name", err.Error())
		return
	}

	instantiateParams := struct {
		Name          string
		ParentNetwork string
		Source        string
	}{
		projectUID,
		fmt.Sprintf("https://vcd-pod-charlie.swisscomcloud.com/api/network/%s", os.Getenv("NETWORK_ID")),
		fmt.Sprintf("https://vcloud.example.com/api/vAppTemplate/vappTemplate-%s", os.Getenv("TEMPLATE_ID")),
	}
	t, err := template.New("instantiateArg").Parse(`<?xml version="1.0" encoding="UTF-8"?> 
	<vcloud:InstantiateVAppTemplateParams 
		 xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1" 
		 xmlns:vcloud="http://www.vmware.com/vcloud/v1.5" 
		 deploy="false" 
		 name="{{ .Name}}" 
		 powerOn="false"> 
		 <vcloud:Description>VApp Description</vcloud:Description> 
		 <vcloud:InstantiationParams> 
			 <vcloud:NetworkConfigSection> 
				 <ovf:Info>The configuration parameters for logical networks</ovf:Info>
				 <vcloud:NetworkConfig 
					 networkName="VM Network"> 
					 <vcloud:Configuration> 
						 <vcloud:ParentNetwork 
							 href="{{ .ParentNetwork}}" 
							 name="VM Network" 
							 type="application/vnd.vmware.vcloud.network+xml"/> 
						 <vcloud:FenceMode>bridged</vcloud:FenceMode> 
					 </vcloud:Configuration> 
				 </vcloud:NetworkConfig> 
			 </vcloud:NetworkConfigSection> 
		 </vcloud:InstantiationParams> 
		 <vcloud:Source 
			 href="{{.Source}}" 
			 name="imported" 
			 type="application/vnd.vmware.vcloud.vAppTemplate+xml"/> 
	</vcloud:InstantiateVAppTemplateParams>`)
	if err != nil {
		handleError("failed to create template: ", err.Error())
		return
	}
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		err = t.Execute(pw, instantiateParams)
		if err != nil {
			handleError("failed to execute template: ", err.Error())
			return
		}
	}()

	url := fmt.Sprintf("https://vcd-pod-charlie.swisscomcloud.com/api/vdc/%s/action/instantiateVAppTemplate", os.Getenv("VDC_ID"))
	req, err := http.NewRequest("POST", url, pr)
	if err != nil {
		handleError("failed to create request: ", err.Error())
		return
	}
	req.Header.Set("x-vcloud-authorization", token)
	req.Header.Set("Accept", "application/*+xml;version=31.0")
	req.Header.Set("Content-Type", "application/vnd.vmware.vcloud.instantiateVAppTemplateParams+xml")

	client := &http.Client{}
	log.Lvlf1("sending POST request: %s", url)
	tef.FlushTaskEventInfo("sending Post request", url)

	resp, err := client.Do(req)
	if err != nil {
		handleError("failed to send request: ", err.Error())
		return
	}
	defer resp.Body.Close()

	log.Lvlf1("got back this status code: %s", resp.Status)
	tef.FlushTaskEventInfo("reading status code", resp.Status)

	if resp.StatusCode != 201 {
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
		if vcloudError.MinorErrorCode != "DUPLICATE_NAME" {
			handleError("unexpected status code", "expected 201, "+
				"got %s with the following message: %s", resp.Status, vcloudError.Message)
			return
		}
		// If the name is already used, that means the eproject should have
		// already been created. If not we return an error
		project, found = models.EProjectList[projectInstID]
		if !found || project == nil {
			handleError("EProject not found", "EProject not found at this instanceID: "+projectInstID)
			return
		}

		log.Lvl1("here is the vapp Href: " + project.EnclaveHref)
		resp, err = getVapp(project.EnclaveHref, token)
		if err != nil {
			handleError("failed to get vApp: ", err.Error())
			project.Status = models.EProjectStatusSetupErrored
			return
		}
		defer resp.Body.Close()

		log.Lvlf1("got back this status code: %s", resp.Status)
		tef.FlushTaskEventInfo("reading status code", resp.Status)

		if resp.StatusCode != 200 {
			handleError("error: expected status code 200", resp.Status)
			project.Status = models.EProjectStatusSetupErrored
			return
		}

		bodyBuf, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			handleError("failed to read body: ", err.Error())
			project.Status = models.EProjectStatusSetupErrored
			return
		}

		log.Lvlf3("got back this body:\n%s", string(bodyBuf))

		err = xml.Unmarshal(bodyBuf, &vapp)
		if err != nil {
			handleError("failed to unmashal body: ", err.Error())
			project.Status = models.EProjectStatusSetupErrored
			return
		}
		// Forgive me for this sin. In this case the vApp has already been
		// created, so we can directly jump to set its properties.
		goto setproperty
	}

	bodyBuf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError("failed to read body: ", err.Error())
		return
	}

	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	err = xml.Unmarshal(bodyBuf, &vapp)
	if err != nil {
		handleError("failed to unmashal body: ", err.Error())
		return
	}

	// Create Project
	project = &models.EProject{
		ProjectUID:    projectUID,
		InstanceID:    projectInstID,
		CloudEndpoint: projectUID,
		EnclaveName:   projectUID,
		EnclaveHref:   vapp.Href,
		Status:        models.EProjectStatusBootingEnclave,
	}
	models.EProjectList[projectInstID] = project

	// ------------------------------------------------------------------------
	// WAIT FOR NEW VAPP TASK TO END

	taskURL = vapp.Tasks.Task.Href
	log.Lvlf1("waiting for the POST vapp task to succeed: %s", taskURL)
	tef.FlushTaskEventInfo("waiting for the POST vapp task to succeed", taskURL)

	err = pollTask(tef, taskURL, token, 10, time.Duration(time.Second*5), 1.3)
	if err != nil {
		handleError("failed to poll the create app task: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	tef.FlushTaskEventImportantInfo("Vm created", "")

	// ------------------------------------------------------------------------
	// SET PROPERTY ON VAPP VM
setproperty:

	vmLink := vapp.Children.VM.Href

	log.Lvlf1("setting property of vm with link: %s", vmLink)
	tef.FlushTaskEventInfo("setting property of vm", vmLink)

	endpointBuf := make([]byte, 8)
	_, err = rand.Read(endpointBuf)
	if err != nil {
		handleError("failed to create random cloud endpoint: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	propertyParam := struct {
		Endpoint     string
		RequestIndex string
	}{
		project.CloudEndpoint,
		requestIndex,
	}

	t, err = template.New("propertyArg").Parse(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
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
			<ns2:Property ns2:userConfigurable="false" ns2:type="string" ns2:password="false" ns2:key="vm_endpoint" ns2:value="{{ .Endpoint}}"/>
			<ns2:Property ns2:userConfigurable="false" ns2:type="string" ns2:password="false" ns2:key="request_index" ns2:value="{{ .RequestIndex}}"/>
		</ns2:ProductSection>
	</ns6:ProductSectionList>
	`)
	if err != nil {
		handleError("failed to create template: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	pr, pw = io.Pipe()

	go func() {
		defer pw.Close()

		err = t.Execute(pw, propertyParam)
		if err != nil {
			handleError("failed to execute template: ", err.Error())
			project.Status = models.EProjectStatusSetupErrored
			return
		}
	}()

	url = fmt.Sprintf("%s/productSections", vmLink)
	req, err = http.NewRequest("PUT", url, pr)
	if err != nil {
		handleError("failed to create request: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	req.Header.Set("x-vcloud-authorization", token)
	req.Header.Set("Accept", "application/*+xml;version=31.0")
	req.Header.Set("Content-Type", "application/vnd.vmware.vcloud.productSections+xml")

	client = &http.Client{}
	log.LLvlf1("sending PUT request to update properties with url: %s", url)
	tef.FlushTaskEventInfo("sending PUT request to update properties", url)

	resp, err = client.Do(req)
	if err != nil {
		handleError("failed to send request: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	defer resp.Body.Close()

	log.Lvlf1("got back this status code: %s", resp.Status)
	tef.FlushTaskEventInfo("reading status code", resp.Status)

	if resp.StatusCode != 202 {
		project.Status = models.EProjectStatusSetupErrored
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
			"got %s with the following message: %s", resp.Status,
			vcloudError.Message)
		return
	}

	bodyBuf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError("failed to read body: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	var task models.Task
	err = xml.Unmarshal(bodyBuf, &task)
	if err != nil {
		handleError("failed to unmashal body: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	// ------------------------------------------------------------------------
	// WAIT FOR NEW PROPERTY TASK TO END

	taskURL = task.Href
	log.Lvlf1("waiting for the PUT properties task to succeed: %s", taskURL)
	tef.FlushTaskEventInfo("waiting for the PUT properties task to succeed", taskURL)

	err = pollTask(tef, taskURL, token, 10, time.Duration(time.Second*5), 1.2)
	if err != nil {
		handleError("failed to poll the set property task: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	tef.FlushTaskEventImportantInfo("Property (cloud endpoint to use) on Vm set", taskURL)

	// ------------------------------------------------------------------------
	// SET THE NETWORK ADAPTER ON VAPP VM

	vmLink = vapp.Children.VM.Href

	log.Lvlf1("updating the VM network adapter with this link: %s", vmLink)
	tef.FlushTaskEventInfo("updating the VM network adapter", vmLink)

	networkParam := struct {
		VMLink    string
		NetworkID string
	}{
		vmLink,
		conf.NetworkID,
	}

	t, err = template.New("networkArg").Parse(`<?xml version="1.0" encoding="UTF-8"?> 
	<vcloud:NetworkConnectionSection 
		 xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1" 
		 xmlns:vcloud="http://www.vmware.com/vcloud/v1.5" 
		 href="{{.VMLink}}" 
		 ovf:required="false" 
		 type="application/vnd.vmware.vcloud.networkConnectionSection+xml"> 
		 <ovf:Info> Specifies the available VM network connections </ovf:Info> 
		 <vcloud:PrimaryNetworkConnectionIndex>0</vcloud:PrimaryNetworkConnectionIndex> 
		 <vcloud:NetworkConnection 
			 needsCustomization="true" 
			 network="{{.NetworkID}}"> 
			 <vcloud:NetworkConnectionIndex>0</vcloud:NetworkConnectionIndex>
			 <vcloud:IpAddress></vcloud:IpAddress>
			 <vcloud:IsConnected>true</vcloud:IsConnected> 
			 <vcloud:IpAddressAllocationMode>POOL</vcloud:IpAddressAllocationMode>
			 <vcloud:NetworkAdapterType>VMXNET3</vcloud:NetworkAdapterType>
		 </vcloud:NetworkConnection> 
		 <vcloud:Link 
			 href="{{.VMLink}}" 
			 rel="edit" 
			 type="application/vnd.vmware.vcloud.networkConnectionSection+xml"/> 
	</vcloud:NetworkConnectionSection>
	`)
	if err != nil {
		handleError("failed to create network connection "+
			"section template: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	pr, pw = io.Pipe()

	go func() {
		defer pw.Close()

		err = t.Execute(pw, networkParam)
		if err != nil {
			handleError("failed to execute template: ", err.Error())
			project.Status = models.EProjectStatusSetupErrored
			return
		}
	}()

	url = fmt.Sprintf("%s/networkConnectionSection", vmLink)
	req, err = http.NewRequest("PUT", url, pr)
	if err != nil {
		handleError("failed to reate request: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	req.Header.Set("x-vcloud-authorization", token)
	req.Header.Set("Accept", "application/*+xml;version=31.0")
	req.Header.Set("Content-Type", "application/vnd.vmware.vcloud.networkConnectionSection+xml")

	client = &http.Client{}

	log.Lvlf1("sending the PUT request to update the adapter: %s", url)
	tef.FlushTaskEventInfo("sending the PUT request to update the network adapter", url)

	resp, err = client.Do(req)
	if err != nil {
		handleError("failed to semd request: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	defer resp.Body.Close()

	log.Lvlf1("got back this status code: %s", resp.Status)
	tef.FlushTaskEventInfo("reading status code", resp.Status)

	if resp.StatusCode != 202 {
		project.Status = models.EProjectStatusSetupErrored
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
		handleError("failed to read body: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	err = xml.Unmarshal(bodyBuf, &task)
	if err != nil {
		handleError("failed to unmashal body: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	// ------------------------------------------------------------------------
	// WAIT FOR THE NETWORKCONNECTIONSECTION TASK TO END

	taskURL = task.Href
	log.Lvlf1("waiting for the PUT networkConnectionSection task to succeed: %s", taskURL)
	tef.FlushTaskEventInfo("waiting for the PUT networkConnectionSection task to succeed", taskURL)

	err = pollTask(tef, taskURL, token, 10, time.Duration(time.Second*5), 1.2)
	if err != nil {
		handleError("failed to poll the set network settings task: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	tef.FlushTaskEventImportantInfo("Network setting on Vm set", taskURL)

	// ------------------------------------------------------------------------
	// GET THE NETWORK IP THAT WAS ALLOCATED

	url = fmt.Sprintf("%s/networkConnectionSection", vmLink)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		handleError("failed to create request: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	req.Header.Set("x-vcloud-authorization", token)
	req.Header.Set("Accept", "application/*+xml;version=31.0")

	client = &http.Client{}

	log.Lvlf1("sending the GET request to fetch the IP: %s", url)
	tef.FlushTaskEventInfo("sending the GET request to fetch the IP",
		fmt.Sprintf("Request URL: %s", url))

	resp, err = client.Do(req)
	if err != nil {
		handleError("failed to send request: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	defer resp.Body.Close()

	log.Lvlf1("got back this status code: %s", resp.Status)
	tef.FlushTaskEventInfo("reading status code", resp.Status)

	if resp.StatusCode != 200 {
		project.Status = models.EProjectStatusSetupErrored
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
		handleError("failed to read body: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	var networkConnectionSection models.NetworkConnectionSection
	err = xml.Unmarshal(bodyBuf, &networkConnectionSection)
	if err != nil {
		handleError("failed to unmashal body: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	ipAddr := networkConnectionSection.NetworkConnection.IPAddress
	project.IPAddr = net.ParseIP(ipAddr)
	if project.IPAddr == nil {
		handleError("error parsing IP Address",
			fmt.Sprintf("Got IPAddress: %s", ipAddr))
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	// ------------------------------------------------------------------------
	// UPDATE THE PROJECT CONTRACT WITH THE ENCLAVE'S IP

	cmd := exec.Command("./pcadmin", "-c", conf.ConfigPath, "contract",
		"project", "invoke", "setURL", "-bc", conf.BCPath, "-sign", conf.KeyID,
		"-darc", conf.DarcID, "-enclaveURL", project.IPAddr.String(), "-i", project.InstanceID)
	tef.FlushTaskEventInfof("setting the pub key of the enclave on the contract",
		"running %s", cmd.Args)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err = cmd.Run()
	if err != nil {
		handleError("failed to set the enclave IP, ",
			"failed to run the command: %s - Output: %s - Err: %s", err.Error(),
			outb.String(), errb.String())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	// ------------------------------------------------------------------------
	// GET THE NEWLY CREATED VAPP

	log.Lvl1("getting the newly created vApp")
	tef.FlushTaskEventInfo("getting the newly created vApp", vapp.Href)

	resp, err = getVapp(vapp.Href, token)
	if err != nil {
		handleError("failed to get vApp: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	defer resp.Body.Close()

	log.Lvlf1("got back this status code: %s", resp.Status)
	tef.FlushTaskEventInfo("reading status code", resp.Status)

	if resp.StatusCode != 200 {
		project.Status = models.EProjectStatusSetupErrored
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
		handleError("unexpected status code", "expected 200, "+
			"got %s with the following message: %s", resp.Status, vcloudError.Message)
		return
	}

	bodyBuf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError("failed to read body: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	err = xml.Unmarshal(bodyBuf, &vapp)
	if err != nil {
		handleError("failed to unmashal body: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	// 3 = suspended
	// 4 = powered on
	// 8 = powered off
	// See the vCloud director documentation for more information
	if vapp.Status != "8" {
		handleError("wrong status", "the newly created vApp should "+
			"have status '8', but we found instead: "+vapp.Status)
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	// ------------------------------------------------------------------------
	// POWER ON THE NEWLY CREATED VAPP

	log.Lvl1("powering on the newly created vApp")
	tef.FlushTaskEventInfo("powering on the newly created vApp", vapp.Href)

	url = fmt.Sprintf("%s/power/action/powerOn", vapp.Href)
	req, err = http.NewRequest("POST", url, nil)
	if err != nil {
		handleError("failed to create request: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	req.Header.Set("x-vcloud-authorization", token)
	req.Header.Set("Accept", "application/*+xml;version=31.0")

	client = &http.Client{}
	log.Lvlf1("sending the POST request to power on the vApp: %s", url)
	tef.FlushTaskEventInfo("sending the POST request to power on the vApp", url)

	resp, err = client.Do(req)
	if err != nil {
		handleError("failed to send request: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	defer resp.Body.Close()

	log.Lvlf1("got back this status code: %s", resp.Status)
	tef.FlushTaskEventInfo("reading status code", resp.Status)

	if resp.StatusCode != 202 {
		project.Status = models.EProjectStatusSetupErrored
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
		handleError("failed to read body: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	err = xml.Unmarshal(bodyBuf, &task)
	if err != nil {
		handleError("failed to unmashal body: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	// ------------------------------------------------------------------------
	// WAIT FOR THE POWER ON TASK TO END

	taskURL = task.Href
	log.Lvlf1("waiting for the POST power on task to succeed: %s", taskURL)
	tef.FlushTaskEventInfo("waiting for the POST power on task to succeed", taskURL)

	err = pollTask(tef, taskURL, token, 10, time.Duration(time.Second*5), 1.3)
	if err != nil {
		handleError("failed to poll the power on task",
			"failed to poll the power on task. Maybe the IP adress %s is "+
				"already used?\nError: %s", project.IPAddr.String(), err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	tef.FlushTaskEventImportantInfo("Vm powered on", "")

	// ------------------------------------------------------------------------
	// WAIT FOR THE CLOUD ENDPOINT TO BE CREATED BY THE VAPP

	log.Lvl1("polling the endpoint that should be created by the vApp")
	tef.FlushTaskEventInfo("polling the endpoint that should be created by "+
		"the vApp", project.CloudEndpoint)

	found = false
	retry := 10
	for retry > 0 {
		found, err = minioClient.BucketExists(project.CloudEndpoint)
		if err != nil {
			handleError("failed to check if the bucket exists", err.Error())
			project.Status = models.EProjectStatusSetupErrored
			return
		}
		if found {
			tef.FlushTaskEventInfo("bucket found", project.CloudEndpoint)
			break
		}

		retry--
		if retry == 0 {
			handleError("failed to find the bucket",
				"failed to find the bucket, timeout reached %s", project.CloudEndpoint)
			project.Status = models.EProjectStatusSetupErrored
			return
		}

		log.Lvlf1("waiting a bit, bucket not found at: %s/%s "+
			"(remaining tries: %d)", os.Getenv("MINIO_ENDPOINT"), project.CloudEndpoint, retry)
		tef.FlushTaskEventInfof("waiting a bit, bucket not found",
			"waiting a bit, bucket not found at: %s/%s (remaining tries: %d)",
			os.Getenv("MINIO_ENDPOINT"), project.CloudEndpoint, retry)
		time.Sleep(time.Second * 10)
	}

	log.LLvl1("getting and checking the content of 'pub_key.txt'")
	tef.FlushTaskEventInfo("checking the pub key", "getting and checking the content of 'pub_key.txt'")

	// Due to some latency, the file can take a bit of time to be readable, so
	// we poll it. From what we observed, it can take sometimes up to 1m40
	// before the startup script is executed. We saw that the startup script is
	// first executed for a short time, then executed a second time (this time
	// wihtout being interrupted) after a while.
	retry = 20
	var pubkeyReader *minio.Object
	for retry > 0 {
		retry--
		pubkeyReader, err = minioClient.GetObject(project.CloudEndpoint, "pub_key.txt", minio.GetObjectOptions{})
		if err != nil {
			handleError("failed to check if 'pub_key.txt' exists: ", err.Error())
			project.Status = models.EProjectStatusSetupErrored
			return
		}
		defer pubkeyReader.Close()

		// If the object doesn't exist we won't get an error until we try to read.
		_, err = pubkeyReader.Stat()
		if err == nil {
			log.LLvl1("'pub_key.txt' found")
			tef.FlushTaskEventInfo("'pub_key.txt' found", project.CloudEndpoint+"/pub_key.txt")
			break
		}

		if retry == 0 {
			handleError("failed to find the public key", "failed to find 'pub_key.txt' in "+project.CloudEndpoint)
			project.Status = models.EProjectStatusSetupErrored
			return
		}

		log.Lvlf1("failed to read the pub key, waiting a bit (remaining tries: %d)", retry)
		tef.FlushTaskEventInfof("pub key not found, sleeping", "failed to read the pub key, waiting 10 seconds (remaining tries: %d)", retry)
		time.Sleep(time.Second * 10)

	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, pubkeyReader)
	if err != nil {
		handleError("failed to copy the content of 'pub_key.txt'", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	result := strings.Trim(buf.String(), "\n\r ")
	keyPattern, err := regexp.Compile("^ed25519:[0-9a-f]{64}$")
	if err != nil {
		handleError("failed to build keyPattern regex", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	if !keyPattern.MatchString(result) {
		handleError("wrong content in pu_key.txt",
			"the content of 'pub_key.txt' is unexpected, found '%s'", result)
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	project.PubKey = result

	log.LLvlf1("ok, erverything is allright (found '%s'), we can power off the enclave", result)
	tef.FlushTaskEventImportantInfo("found the pub key", fmt.Sprintf("ok, everything is allright (found '%s'), we can power off the enclave", result))

	// ------------------------------------------------------------------------
	// UPDATE THE PROJECT CONTRACT WITH THE ENCLAVE'S PUB KEY

	cmd = exec.Command("./pcadmin", "-c", conf.ConfigPath, "contract",
		"project", "invoke", "setEnclavePubKey", "-bc", conf.BCPath, "-sign",
		conf.KeyID, "-darc", conf.DarcID, "-pubKey", project.PubKey, "-i", project.InstanceID)
	tef.FlushTaskEventInfof("setting the pub key of the enclave on the contract",
		"running %s", cmd.Args)

	outb.Reset()
	errb.Reset()
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err = cmd.Run()
	if err != nil {
		handleError("failed to set the enclave pub key, ",
			"failed to run the command: %s - Output: %s - Err: %s", err.Error(),
			outb.String(), errb.String())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	// ------------------------------------------------------------------------
	// SHUTDOWN THE NEWLY CREATED VAPP

	log.Lvl1("shutdown (undeploy) the newly created vApp")
	tef.FlushTaskEventInfo("shut down the emclave", "shut down (undeploy) the newly created vApp")
	// We can use the 'shutdown' command because VMware tools is installed on
	// the VM, otherwise we must use 'powerOff', but shis may leave the enclave
	// with an unstable state (with empty files for example).
	undeployArg := []byte(`<?xml version="1.0" encoding="UTF-8"?>
	<vcloud:UndeployVAppParams
		xmlns:vcloud="http://www.vmware.com/vcloud/v1.5">
		<vcloud:UndeployPowerAction>shutdown</vcloud:UndeployPowerAction> 
	</vcloud:UndeployVAppParams>`)

	url = fmt.Sprintf("%s/action/undeploy", vapp.Href)
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(undeployArg))
	if err != nil {
		handleError("failed to create request", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	req.Header.Set("x-vcloud-authorization", token)
	req.Header.Set("Accept", "application/*+xml;version=31.0")
	req.Header.Set("Content-Type", "application/vnd.vmware.vcloud.undeployVAppParams+xml")

	client = &http.Client{}
	log.Lvlf1("sending the POST request to shutdown (undeploy) the vApp: %s", url)
	tef.FlushTaskEventInfo("sending the POST request to shutdown (undeploy) the vApp", url)

	resp, err = client.Do(req)
	if err != nil {
		handleError("failed to send request: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}
	defer resp.Body.Close()

	log.Lvlf1("got back this status code: %s", resp.Status)
	tef.FlushTaskEventInfo("reading status code", resp.Status)

	if resp.StatusCode != 202 {
		project.Status = models.EProjectStatusSetupErrored
		bodyBuf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			handleError("unexpected status code",
				"got an unexpected status code %s and failed to read body:%s ",
				resp.Status, err.Error())
			return
		}
		var vcloudError models.VcloudError
		err = xml.Unmarshal(bodyBuf, &vcloudError)
		if err != nil {
			handleError("unexpected status code",
				"got an unexpected status code %s and failed to unmashall body:%s ",
				resp.Status, err.Error())
			return
		}
		handleError("unexpected status code", "expected 202, "+
			"got %s with the following message: %s", resp.Status, vcloudError.Message)
		return
	}

	bodyBuf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError("failed to read body: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	err = xml.Unmarshal(bodyBuf, &task)
	if err != nil {
		handleError("failed to unmarshal body: ", err.Error())
		project.Status = models.EProjectStatusSetupErrored
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
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	tef.FlushTaskEventImportantInfo("Vm shut down", "")

	// ------------------------------------------------------------------------
	// END OF THE PROCESS

	log.LLvl1("new enclave created, configured, booted, working fine and now powered off")
	project.Status = models.EProjectStatusSetupReady

	tef.FlushTaskEventInfo("updating the contract status",
		"setting the contract instance status to 'preparedOK'")
	err = models.UpdateProjectcStatus(conf, "preparedOK", project.InstanceID)
	if err != nil {
		handleError("failed to update the contract status to 'preparedOK'",
			err.Error())
		project.Status = models.EProjectStatusSetupErrored
		return
	}

	tef.FlushTaskEventCloseOK("enclave successfully set up", "new enclave created, configured, booted, working fine and now powered off")

}

// VappsShowGet gets a single vApp and sends its json representation
func VappsShowGet(w http.ResponseWriter, r *http.Request, gs *sessions.CookieStore) {
	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		helpers.SendRequestError(errors.New("failed to get param"), w)
		return
	}

	token, err := helpers.GetToken(w)
	if err != nil {
		helpers.SendRequestError(err, w)
		return
	}

	url := fmt.Sprintf("https://%s/api/vApp/%s", helpers.VcdHost, id)
	resp, err := getVapp(url, token)
	defer resp.Body.Close()
	if err != nil {
		helpers.SendRequestError(errors.New("failed to get vApp: "+err.Error()), w)
		return
	}

	if resp.StatusCode != 200 {
		helpers.SendRequestError(helpers.NewNotOkError(resp), w)
		return
	}

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		helpers.SendRequestError(errors.New("failed to read body: "+err.Error()), w)
		return
	}
	log.Lvlf3("got back this body:\n%s", string(bodyBuf))

	var vapp models.VApp
	err = xml.Unmarshal(bodyBuf, &vapp)
	if err != nil {
		helpers.SendRequestError(errors.New("failed to unmarshal body: "+err.Error()), w)
		return
	}

	js, err := json.MarshalIndent(vapp, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// getVapp sends a request to get a single vApp and returns the response
func getVapp(url string, token string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.New("failed to build request:" + err.Error())
	}
	req.Header.Set("x-vcloud-authorization", token)
	req.Header.Set("Accept", "application/*+xml;version=31.0")

	client := &http.Client{}
	log.LLvlf1("sending POST request to get vApp: %s", url)
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("failed to send request: " + err.Error())
	}
	return resp, nil
}

// pollTaskk is used when an asynchronous call is made to the vCloud API. When
// doing so, the vCloud API returns a task that indicates the status of the
// request. This function poll the task based on the retry, wait and exponent
// parameters.
func pollTask(tef *xhelpers.TaskEventFFactory, taskURL,
	token string, retry int, wait time.Duration, exponent float64) error {

	log.Lvlf1("starting to poll the task maximum %d times: %s", retry, taskURL)
	tef.FlushTaskEventInfof("starting to poll", "starting to poll the task maximum %d times: %s", retry, taskURL)

	var task models.Task
	for retry > 0 {
		retry--

		// Poll the task
		req, err := http.NewRequest("GET", taskURL, nil)
		if err != nil {
			return errors.New("failed to build request for the task:" + err.Error())
		}
		req.Header.Set("x-vcloud-authorization", token)
		req.Header.Set("Accept", "application/*;version=31.0")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return errors.New("failed to send request: " + err.Error())
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			bodyBuf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("got an unexpected status code %s "+
					"and failed to read body:%s ", resp.Status, err.Error())
			}
			var vcloudError models.VcloudError
			err = xml.Unmarshal(bodyBuf, &vcloudError)
			if err != nil {
				return fmt.Errorf("got an unexpected status code %s "+
					"and failed to unmashall body:%s ", resp.Status, err.Error())
			}
			return fmt.Errorf("expected 200, got %s with the "+
				"following message: %s", resp.Status, vcloudError.Message)
		}

		bodyBuf2, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.New("failed to read body: " + err.Error())
		}

		err = xml.Unmarshal(bodyBuf2, &task)
		if err != nil {
			return errors.New("failed to unmarshal body: " + err.Error())
		}

		if task.Status == "success" {
			log.Lvl1("Task succeded")
			tef.FlushTaskEventInfof("task succeded", task.EndTime, task.Operation)
			break
		}

		if retry == 0 {
			return errors.New("failed to poll the task, no retry left (retry==0)")
		}

		if task.Status == "queued" || task.Status == "running" {
			log.Lvlf1("Sleeping for %fs: status is %s, progress is %s "+
				"(remaining tries: %d)", wait.Seconds(), task.Status, task.Progress, retry)
			tef.FlushTaskEventInfof("sleeping", "Sleeping for %fs: status is %s, progress is %s "+
				"(remaining tries: %d)", wait.Seconds(), task.Status, task.Progress, retry)
			time.Sleep(wait)
			wait = time.Duration(float64(wait.Nanoseconds()) * exponent)
		} else {
			return fmt.Errorf("the task got an unexpected status: %s - %s - %s", task.Status, task.Details, task.Error.Message)
		}
	}
	return nil
}
