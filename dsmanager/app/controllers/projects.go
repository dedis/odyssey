package controllers

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/dedis/odyssey/catalogc"
	"github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/dedis/odyssey/dsmanager/app/models"
	"github.com/dedis/odyssey/projectc"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/protobuf"
)

// ProjectsIndexHandler ...
func ProjectsIndexHandler(store *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projectsIndexGet(w, r, store, conf)
		case http.MethodPost:
			projectsPost(w, r, store, conf)
		default:
			log.Error("Only post and get are allowed for projects")
		}
	}
}

// ProjectsShowHandler ...
func ProjectsShowHandler(gs *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projectsShowGet(w, r, gs, conf)
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				helpers.RedirectWithErrorFlash(r.URL.String(), "failed to read form", w, r, gs)
				return
			}
			switch r.PostFormValue("_method") {
			case "put":
				projectsShowPut(w, r, gs, conf)
			case "delete":
				projectsShowDelete(w, r, gs, conf)
			default:
				helpers.RedirectWithErrorFlash(r.URL.String(), "only PUT and DELETE allowed", w, r, gs)
			}
		}
	}
}

// ProjectsShowAttributesHandler ...
func ProjectsShowAttributesHandler(gs *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projectsShowAttributesGet(w, r, gs, conf)
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				helpers.RedirectWithErrorFlash(r.URL.String(), "failed to read form", w, r, gs)
				return
			}
			switch r.PostFormValue("_method") {
			case "put":
				projectsShowAttributesPut(w, r, gs, conf)
			default:
				helpers.RedirectWithErrorFlash(r.URL.String(), "only PUT allowed", w, r, gs)
			}
		}
	}
}

// ProjectsShowEnclaveHandler ...
func ProjectsShowEnclaveHandler(gs *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projectsShowEnclaveGet(w, r, gs, conf)
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				helpers.RedirectWithErrorFlash(r.URL.String(), "failed to read form", w, r, gs)
				return
			}
			switch r.PostFormValue("_method") {
			case "delete":
				projectsShowEnclaveDelete(w, r, gs, conf)
			default:
				helpers.RedirectWithErrorFlash(r.URL.String(), "only DELETE allowed", w, r, gs)
			}
		default:
			helpers.RedirectWithErrorFlash("/", "only GET and DELETE allowed", w, r, gs)
		}
	}
}

// ProjectsShowUnlockHandler ...
func ProjectsShowUnlockHandler(gs *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				helpers.RedirectWithErrorFlash(r.URL.String(), "failed to read form", w, r, gs)
				return
			}
			switch r.PostFormValue("_method") {
			case "put":
				projectsShowUnlockPut(w, r, gs, conf)
			default:
				helpers.RedirectWithErrorFlash(r.URL.String(), "only PUT allowed", w, r, gs)
			}
		default:
			helpers.RedirectWithErrorFlash("/", "only PUT allowed", w, r, gs)
		}
	}
}

// ProjectsShowDebugHandler ...
func ProjectsShowDebugHandler(gs *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projectShowDebugGet(w, r, gs, conf)
		}
	}
}

// ProjectsShowStatusHandler ...
func ProjectsShowStatusHandler(gs *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				helpers.RedirectWithErrorFlash(r.URL.String(), "failed to read form", w, r, gs)
				return
			}
			switch r.PostFormValue("_method") {
			case "put":
				projectsShowStatusPut(w, r, gs, conf)
			default:
				helpers.RedirectWithErrorFlash(r.URL.String(), "only PUT allowed", w, r, gs)
			}
		default:
			helpers.RedirectWithErrorFlash("/", "only PUT allowed", w, r, gs)
		}
	}
}

// ProjectsShowStatusStreamHandler ...
func ProjectsShowStatusStreamHandler(gs *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projectsShowStatusStreamGet(w, r, gs, conf)
		}
	}
}

// ProjectsRequestsShowHandler ...
func ProjectsRequestsShowHandler(gs *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projectsRequestShowGet(w, r, gs, conf)
		}
	}
}

// ProjectsRequestsShowCloudstreamHandler ...
func ProjectsRequestsShowCloudstreamHandler(gs *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projectsRequestShowCloudstreamGet(w, r, gs, conf)
		}
	}
}

// ProjectsRequestsShowStatusStreamHandler ...
func ProjectsRequestsShowStatusStreamHandler(gs *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projectsRequestShowStatusStreamGet(w, r, gs, conf)
		}
	}
}

// ProjectsRequestsTasksShowStreamHandler ...
func ProjectsRequestsTasksShowStreamHandler(gs *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projectsRequestTasksShowStream(w, r, gs, conf)
		}
	}
}

// ProjectsRequestsTasksShowDebugHandler ...
func ProjectsRequestsTasksShowDebugHandler(gs *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projectsRequestTasksShowDebugGet(w, r, gs, conf)
		}
	}
}

// ProjectsRequestsTasksShowStatusHandler ...
func ProjectsRequestsTasksShowStatusHandler(gs *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				helpers.RedirectWithErrorFlash(r.URL.String(), "failed to read form", w, r, gs)
				return
			}
			switch r.PostFormValue("_method") {
			case "put":
				projectsRequestTasksShowStatusPut(w, r, gs, conf)
			default:
				helpers.RedirectWithErrorFlash(r.URL.String(), "only PUT allowed", w, r, gs)
			}
		default:
			helpers.RedirectWithErrorFlash("/", "only PUT allowed", w, r, gs)
		}
	}
}

// @Summary Create a new project
// @Description This will initialize, create and store a new project.
// @Produce  html
// @Accept multipart/form-data
// @Param datasetIDs formData string true "list of dataset IDs separated by commas"
// @Router /projects [post]
func projectsPost(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	type viewData struct {
		Title    string
		Datasets []models.Dataset
		Flash    []helpers.Flash
	}

	// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
	err := r.ParseForm()
	if err != nil {
		helpers.RedirectWithErrorFlash("/datasets", "failed to parse POST "+
			"argument:\n"+err.Error(), w, r, store)
		return
	}
	log.Lvlf1("We got this post form: %v", r.PostForm)

	datasetIDs, ok := r.PostForm["datasetIDs"]
	if !ok || len(datasetIDs) == 0 {
		helpers.RedirectWithErrorFlash("/datasets", "the form key 'datasetIDs' is empty, "+
			"please select at least one dataset", w, r, store)
		return
	}

	project := models.NewProject("", "")

	go func() {
		request, task := project.RequestCreateProjectInstance(datasetIDs, conf)
		// if either ones are nil that means an error happened and we must
		// abort.
		if request == nil || task == nil {
			return
		}
		project.RequestBootEnclave(request, task, conf)
	}()

	helpers.RedirectWithInfoFlash("/projects",
		fmt.Sprintf("A new project with uniq id '%s' and title '%s' has been "+
			"created and the request to set up an enclave submitted",
			project.UID, project.Title), w, r, store)
	return
}

// @Summary Gets the list of projects
// @Description Return every project found in the DB.
// @Produce  html
// @Router /projects [get]
func projectsIndexGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	t, err := template.ParseFiles("views/layout.gohtml", "views/projects/index.gohtml")
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	type viewData struct {
		Title    string
		Projects []*models.Project
		Flash    []helpers.Flash
	}

	flashes, err := helpers.ExtractFlash(w, r, store)
	if err != nil {
		log.Errorf("Failed to get flash: %s\n", err.Error())
	}

	projectSlice := []*models.Project{}
	for _, value := range models.ProjectList {
		projectSlice = append(projectSlice, value)
	}

	sort.Sort(sort.Reverse(models.ProjectSorter(projectSlice)))

	p := &viewData{
		Title:    "List of projects",
		Flash:    flashes,
		Projects: projectSlice,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		log.Errorf("Error while executing template: %s\n", err.Error())
	}
}

func projectsShowGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get the project id in url", w, r, store)
		return
	}

	t, err := template.New("template").Funcs(template.FuncMap{
		"isGreen": func(index int, status models.ProjectStatus) string {
			switch index {
			case 1:
				if status == models.ProjectStatusPreparingEnclaveDone ||
					status == models.ProjectStatusUpdatingAttributes ||
					status == models.ProjectStatusAttributesUpdatedDone ||
					status == models.ProjectStatusAttributesUpdatedErrored ||
					status == models.ProjectStatusUnlockingEnclave ||
					status == models.ProjectStatusUnlockingEnclaveDone ||
					status == models.ProjectStatusUnlockingEnclaveErrored ||
					status == models.ProjectStatusDeletingEnclave ||
					status == models.ProjectStatusDeletingEnclaveDone ||
					status == models.ProjectStatusDeletingEnclaveErrored {
					return "green"
				}
			case 2:
				if status == models.ProjectStatusAttributesUpdatedDone ||
					status == models.ProjectStatusUnlockingEnclave ||
					status == models.ProjectStatusUnlockingEnclaveDone ||
					status == models.ProjectStatusUnlockingEnclaveErrored ||
					status == models.ProjectStatusDeletingEnclave ||
					status == models.ProjectStatusDeletingEnclaveDone ||
					status == models.ProjectStatusDeletingEnclaveErrored {
					return "green"
				}
			case 3:
				if status == models.ProjectStatusUnlockingEnclaveDone ||
					status == models.ProjectStatusDeletingEnclave ||
					status == models.ProjectStatusDeletingEnclaveDone ||
					status == models.ProjectStatusDeletingEnclaveErrored {
					return "green"
				}
			case 4:
				if status == models.ProjectStatusDeletingEnclaveDone {
					return "green"
				}
			}
			return ""
		},
		"isOrange": func(index int, status models.ProjectStatus) string {
			switch index {
			case 1:
				if status == models.ProjectStatusPreparingEnclave {
					return "orange"
				}
			case 2:
				if status == models.ProjectStatusUpdatingAttributes {
					return "orange"
				}
			case 3:
				if status == models.ProjectStatusUnlockingEnclave {
					return "orange"
				}
			case 4:
				if status == models.ProjectStatusDeletingEnclave {
					return "orange"
				}
			}
			return ""
		},
		"isRed": func(index int, status models.ProjectStatus) string {
			switch index {
			case 1:
				if status == models.ProjectStatusPreparingEnclaveErrored {
					return "red"
				}
			case 2:
				if status == models.ProjectStatusAttributesUpdatedErrored {
					return "red"
				}
			case 3:
				if status == models.ProjectStatusUnlockingEnclaveErrored {
					return "red"
				}
			case 4:
				if status == models.ProjectStatusDeletingEnclaveErrored {
					return "red"
				}
			}
			return ""
		},
	}).ParseFiles("views/layout.gohtml", "views/projects/show.gohtml")
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	type viewData struct {
		Title          string
		Project        *models.Project
		Flash          []helpers.Flash
		Datasets       []*catalogc.Dataset
		Purposes       []string
		Uses           []string
		SortedRequests []*models.Request
		// In case the project failed to unlock the enclave, we would like to
		// display the reasons to the user.
		FailedReasons       string
		ProjectContractData *projectc.ProjectData
	}

	project, ok := models.ProjectList[id]
	if !ok || project == nil {
		helpers.RedirectWithErrorFlash("/", "Project not found", w, r, store)
		return
	}

	failedReason := ""
	if project.Status == models.ProjectStatusUnlockingEnclaveErrored {
		latestMsg, latestDetails := project.GetLastestTaskMsg()
		log.Info("latest message: ", latestMsg, "latest details: ", latestDetails)
		lastI := strings.LastIndex(latestDetails, "attr:allowed verification failed")
		if lastI != -1 {
			failedReason = latestDetails[lastI:len(latestDetails)]
		} else {
			lastI = strings.LastIndex(latestDetails, "attr:must_have verification failed")
			if lastI != -1 {
				failedReason = latestDetails[lastI:len(latestDetails)]
			}
		}
	}

	var datasets []*catalogc.Dataset
	var projectContractData *projectc.ProjectData
	if project.InstanceID != "" {
		cmd := exec.Command("./pcadmin", "contract", "project", "get", "-i",
			project.InstanceID, "-bc", conf.BCPath, "-x")
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			helpers.RedirectWithErrorFlash("/projects", fmt.Sprintf(
				"failed to get the project instance: %s - Output: %s - Err: %s",
				err.Error(), outb.String(), errb.String()), w, r, store)
			return
		}

		cmdOut := outb.Bytes()
		projectContractData = &projectc.ProjectData{}
		err = protobuf.Decode(cmdOut, projectContractData)
		if err != nil {
			helpers.RedirectWithErrorFlash("/projects", "failed to decode "+
				"project innstance: "+err.Error(), w, r, store)
			return
		}

		datasets = make([]*catalogc.Dataset, len(projectContractData.Datasets))
		// instanceID is the calypsoWriteID
		for i, instanceID := range projectContractData.Datasets {
			log.LLvl1("parsing instanceID", instanceID)
			cmd := exec.Command("./catadmin", "contract", "catalog", "getSingleDataset", "-calypsoWriteID",
				instanceID.String(), "-i", conf.CatalogID, "-bc", conf.BCPath, "--export")
			cmd.Stderr = &errb
			outBuf, err := cmd.Output()
			if err != nil {
				helpers.RedirectWithErrorFlash("/projects", fmt.Sprintf(
					"failed to get the write instance: %s - Output: %s - Err: %s",
					err.Error(), outb.String(), errb.String()), w, r, store)
				return
			}

			var dataset catalogc.Dataset
			err = protobuf.Decode(outBuf, &dataset)
			if err != nil {
				helpers.RedirectWithErrorFlash("/", "failed to decode dataset "+
					"struct: "+err.Error(), w, r, store)
			}
			datasets[i] = &dataset
		}
	} else {
		helpers.AddFlash(w, r, "project instance ID not found, you may "+
			"reload this page later to see the list of datasets", store, helpers.Warning)
	}

	flashes, err := helpers.ExtractFlash(w, r, store)
	if err != nil {
		log.Errorf("Failed to get flash: %s\n", err.Error())
	}

	// We make a copy of the request slice to have it sorted. This should not be
	// too memory consuming since it's only a slice of pointers.
	sortedRequests := make([]*models.Request, len(project.Requests))
	copy(sortedRequests, project.Requests)
	sort.Sort(sort.Reverse(models.RequestSorter(sortedRequests)))
	p := &viewData{
		Title:               "List of projects",
		Flash:               flashes,
		Project:             project,
		Datasets:            datasets,
		SortedRequests:      sortedRequests,
		FailedReasons:       failedReason,
		ProjectContractData: projectContractData,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		log.Errorf("Error while executing template: %s\n", err.Error())
	}
}

// At the moment this method only retries to prepare the enclave. We may further
// want to update the datasets of the project, or the title.
func projectsShowPut(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {
	// we assume that r.ParseForm() has already been called.

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get the project id in url", w, r, store)
		return
	}

	project, ok := models.ProjectList[id]
	if !ok || project == nil {
		helpers.RedirectWithErrorFlash("/", "Project not found", w, r, store)
		return
	}

	if project.InstanceID == "" {
		helpers.RedirectWithErrorFlash("/", "The instance ID on the project is "+
			"not set yet.\nThat means either it is being created and you may "+
			"try this later, or the project creation failed before saving the "+
			"project data.\nIn that case you can't re-use this project and must "+
			"create a new one.", w, r, store)
		return
	}

	cmd := exec.Command("./pcadmin", "contract", "project", "get", "-i",
		project.InstanceID, "-bc", conf.BCPath, "-x")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		helpers.RedirectWithErrorFlash("/projects", fmt.Sprintf(
			"failed to get the project instance: %s - Output: %s - Err: %s",
			err.Error(), outb.String(), errb.String()), w, r, store)
		return
	}

	cmdOut := outb.Bytes()
	projectContractData := &projectc.ProjectData{}
	err = protobuf.Decode(cmdOut, projectContractData)
	if err != nil {
		helpers.RedirectWithErrorFlash("/projects", "failed to decode "+
			"project instance: "+err.Error(), w, r, store)
		return
	}

	datasetIDs := make([]string, len(projectContractData.Datasets))
	for i := range projectContractData.Datasets {
		datasetIDs[i] = projectContractData.Datasets[i].String()
	}

	go func() {
		project.RequestBootEnclave(nil, nil, conf)
	}()

	helpers.RedirectWithInfoFlash("/projects/"+project.UID, "New request to "+
		"prepare the enclave has been submitted", w, r, store)
	return
}

func projectsShowDelete(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {
	// we assume that r.ParseForm() has already been called.

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get the project id in url", w, r, store)
		return
	}

	project, ok := models.ProjectList[id]
	if !ok || project == nil {
		helpers.RedirectWithErrorFlash("/", "Project not found", w, r, store)
		return
	}

	delete(models.ProjectList, id)

	helpers.RedirectWithInfoFlash("/projects", "Project '"+id+
		"' deleted", w, r, store)
	return
}

func projectsShowAttributesGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get the project id in url", w, r, store)
		return
	}

	type viewData struct {
		Title           string
		Project         *models.Project
		Flash           []helpers.Flash
		Metadata        *catalogc.Metadata
		ProjectMetadata *catalogc.Metadata
		Datasets        []*catalogc.Dataset
	}

	flashes, err := helpers.ExtractFlash(w, r, store)
	if err != nil {
		log.Errorf("Failed to get flash: %s\n", err.Error())
	}

	// Let's get the project instance

	project, ok := models.ProjectList[id]
	if !ok || project == nil {
		helpers.RedirectWithErrorFlash("/", "Project not found", w, r, store)
		return
	}

	cmd := exec.Command("./pcadmin", "contract", "project", "get", "-i",
		project.InstanceID, "-bc", conf.BCPath, "-x")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		helpers.RedirectWithErrorFlash("/projects", fmt.Sprintf("failed to "+
			"get the project instance: %s - Output: %s - Err: %s", err.Error(),
			outb.String(), errb.String()), w, r, store)
		return
	}

	cmdOut := outb.Bytes()
	projectContractData := &projectc.ProjectData{}
	err = protobuf.Decode(cmdOut, projectContractData)
	if err != nil {
		helpers.RedirectWithErrorFlash("/projects", "failed to decode "+
			"project innstance: "+err.Error(), w, r, store)
		return
	}

	// For each dataset, we need to get its corresponding metadata stored on the
	// catalog.
	datasetList := []*catalogc.Dataset{}
	for _, calypsoWriteID := range projectContractData.Datasets {
		cmd = new(exec.Cmd)
		cmd = exec.Command("./catadmin", "contract", "catalog",
			"getSingleDataset", "-i", conf.CatalogID, "--calypsoWriteID",
			calypsoWriteID.String(), "--bc", conf.BCPath, "--export")
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))

		outb.Reset()
		errb.Reset()
		cmd.Stdout = &outb
		cmd.Stderr = &errb

		err = cmd.Run()

		dataset := &catalogc.Dataset{}
		err := protobuf.Decode(outb.Bytes(), dataset)
		if err != nil {
			helpers.RedirectWithErrorFlash("/projects", "failed to decode "+
				"dataset: "+err.Error(), w, r, store)
			return
		}

		datasetList = append(datasetList, dataset)
	}

	// Get the catalog

	cmd = new(exec.Cmd)
	cmd = exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
		"catalog", "getMetadata", "-i", conf.CatalogID, "-bc", conf.BCPath,
		"--export")

	log.Info(fmt.Sprintf("command created: %s", cmd.Args))
	outb.Reset()
	errb.Reset()
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		helpers.RedirectWithErrorFlash("/", fmt.Sprintf("failed to get the "+
			"Metadata: %s - Output: %s - Err: %s", err.Error(),
			outb.String(), errb.String()), w, r, store)
		return
	}
	metadata := &catalogc.Metadata{}
	err = protobuf.Decode(outb.Bytes(), metadata)
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "failed to unmarshal result: "+
			err.Error(), w, r, store)
		return
	}
	metadata.Reset()

	t, err := template.New("template").Funcs(template.FuncMap{
		"addCheck": func(els []string, target string) string {
			for _, el := range els {
				if el == target {
					return "checked"
				}
			}
			return ""
		},
	}).ParseFiles("views/layout.gohtml", "views/projects/attributes.gohtml")
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	p := &viewData{
		Title:           "List of projects",
		Flash:           flashes,
		Project:         project,
		Metadata:        metadata,
		ProjectMetadata: projectContractData.Metadata,
		Datasets:        datasetList,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		log.Errorf("Error while executing template: %s\n", err.Error())
	}
}

func projectsShowAttributesPut(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {
	// Here we assume that "r.ParseForm" has already been called

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get the project id in url", w, r, store)
		return
	}

	project, ok := models.ProjectList[id]
	if !ok || project == nil {
		helpers.RedirectWithErrorFlash("/", "Project not found", w, r, store)
		return
	}

	log.Lvlf1("We got this post form: %v", r.PostForm)

	go func() {
		project.RequestUpdateAttributes(r.PostForm, conf)
	}()

	helpers.RedirectWithInfoFlash("/projects/"+project.UID, "Request to "+
		"update the attributes submitted", w, r, store)
	return
}

func projectsShowEnclaveGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	type viewData struct {
		Title   string
		Project *models.Project
		Flash   []helpers.Flash
		URL     string
	}

	t, err := template.ParseFiles("views/layout.gohtml", "views/projects/enclave.gohtml")
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get the project id in url", w, r, store)
		return
	}

	project, ok := models.ProjectList[id]
	if !ok || project == nil {
		helpers.RedirectWithErrorFlash("/", "Project not found", w, r, store)
		return
	}

	cmd := exec.Command("./pcadmin", "contract", "project", "get", "-i",
		project.InstanceID, "-bc", conf.BCPath, "-x")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		helpers.RedirectWithErrorFlash("/projects", fmt.Sprintf("failed to "+
			"get the project instance: %s - Output: %s - Err: %s", err.Error(),
			outb.String(), errb.String()), w, r, store)
		return
	}
	cmdOut := outb.Bytes()
	projectContractData := &projectc.ProjectData{}
	err = protobuf.Decode(cmdOut, projectContractData)
	if err != nil {
		helpers.RedirectWithErrorFlash("/projects", "failed to decode "+
			"project innstance: "+err.Error(), w, r, store)
		return
	}

	enclaveURL := strings.TrimSpace(projectContractData.EnclaveURL)

	flashes, err := helpers.ExtractFlash(w, r, store)
	if err != nil {
		log.Errorf("Failed to get flash: %s\n", err.Error())
	}

	p := &viewData{
		Title:   "List of projects",
		Flash:   flashes,
		Project: project,
		URL:     enclaveURL,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		log.Errorf("Error while executing template: %s\n", err.Error())
	}
}

func projectsShowEnclaveDelete(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {
	// Here we assume that "r.ParseForm" has already been called

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get the project id in url", w, r, store)
		return
	}

	project, ok := models.ProjectList[id]
	if !ok || project == nil {
		helpers.RedirectWithErrorFlash("/", "Project not found", w, r, store)
		return
	}

	log.Lvlf1("We got this post form: %v", r.PostForm)

	go func() {
		project.RequestDeleteEnclave(conf)
	}()

	helpers.RedirectWithInfoFlash("/projects/"+project.UID, "Request to "+
		"delete the enclave submitted", w, r, store)
	return
}

func projectsShowUnlockPut(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {
	// Here we assume that "r.ParseForm" has already been called

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get the project id in url", w, r, store)
		return
	}

	project, ok := models.ProjectList[id]
	if !ok || project == nil {
		helpers.RedirectWithErrorFlash("/", "Project not found", w, r, store)
		return
	}

	log.Lvlf1("We got this post form: %v", r.PostForm)

	go func() {
		project.RequestUnlockEnclave(conf)
	}()

	helpers.RedirectWithInfoFlash("/projects/"+project.UID, "Request to "+
		"unlock the enclave submitted", w, r, store)
	return
}

func projectShowDebugGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	type viewData struct {
		Title   string
		Project *models.Project
		Flash   []helpers.Flash
	}

	t, err := template.ParseFiles("views/layout.gohtml", "views/projects/debug.gohtml")
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get the project id in url", w, r, store)
		return
	}

	project, ok := models.ProjectList[id]
	if !ok || project == nil {
		helpers.RedirectWithErrorFlash("/", "Project not found", w, r, store)
		return
	}

	flashes, err := helpers.ExtractFlash(w, r, store)
	if err != nil {
		log.Errorf("Failed to get flash: %s\n", err.Error())
	}

	p := &viewData{
		Title:   "List of projects",
		Flash:   flashes,
		Project: project,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		log.Errorf("Error while executing template: %s\n", err.Error())
	}
}

func projectsShowStatusPut(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {
	// Here we assume that "r.ParseForm" has already been called

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get the project id in url", w, r, store)
		return
	}

	project, ok := models.ProjectList[id]
	if !ok || project == nil {
		helpers.RedirectWithErrorFlash("/", "Project not found", w, r, store)
		return
	}

	status := r.PostForm.Get("status")
	if status == "" {
		helpers.RedirectWithErrorFlash("/", "Status not found in form", w, r, store)
		return
	}

	project.Status = models.ProjectStatus(status)

	helpers.RedirectWithInfoFlash("/projects/"+project.UID, fmt.Sprintf(
		"Project '%s' updated with status '%s'", project.Title, project.Status), w, r, store)
}

func projectsShowStatusStreamGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Error("the response writer does not support streaming")
		helpers.RedirectWithErrorFlash("/", "the response writer does not "+
			"support streaming", w, r, store)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "false")

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		fmt.Fprintf(w, "data: %s\n\n", "failed to get project id")
		flusher.Flush()
		return
	}

	project, ok := models.ProjectList[id]
	if !ok || project == nil {
		fmt.Fprintf(w, "data: %s\n\n", "project not found")
		return
	}

	statusNotifier := project.StatusNotifier

	// If the status notifier is terminated, then no need for a client
	var client *helpers.StatusNotifierSubscriber
	if !statusNotifier.Terminated {
		client = statusNotifier.Subscribe()
	} else {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for {
		select {
		case status := <-client.NotifyStream:
			fmt.Fprintf(w, "data: %s\n\n", status)
			flusher.Flush()
		default:
			select {
			case <-client.Done:
				// We ensure we empty our Tastream before exiting
			streamFlush:
				for {
					select {
					case status := <-client.NotifyStream:
						fmt.Fprintf(w, "data: %s\n\n", status)
						flusher.Flush()
					default:
						break streamFlush
					}
				}
				log.Info("client done...")
				return
			default:
			}
		}
		time.Sleep(time.Millisecond * 100)
	}

}

func projectsRequestShowGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	params := mux.Vars(r)
	pid := params["pid"]
	if pid == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get the project id in url", w, r, store)
		return
	}

	ridStr := params["rid"]
	if ridStr == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get the request id in url", w, r, store)
		return
	}

	rid, err := strconv.Atoi(ridStr)
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "request id is not an int", w, r, store)
		return
	}

	t, err := template.ParseFiles("views/layout.gohtml", "views/projects/requests/show.gohtml")
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	type viewData struct {
		Title           string
		Project         *models.Project
		Request         *models.Request
		Flash           []helpers.Flash
		RID             int
		CloudLogs       []*helpers.TaskEvent
		CloudLogsStatus helpers.CloudNotifierStatus
	}

	flashes, err := helpers.ExtractFlash(w, r, store)
	if err != nil {
		log.Errorf("Failed to get flash: %s\n", err.Error())
	}

	project, ok := models.ProjectList[pid]
	if !ok || project == nil {
		helpers.RedirectWithErrorFlash("/", "Project not found", w, r, store)
		return
	}

	if len(project.Requests) <= rid || rid < 0 {
		helpers.RedirectWithErrorFlash("/", "request index is out of bound", w, r, store)
		return
	}

	request := project.Requests[rid]

	// Here we take a small short circuit: we get the current saved log in the
	// cloud, then the view will wait for newly created log via the
	// .../cloudstream entry. But there is a small risk that between the time we
	// got the already saved logs and the time we start reading, newly created
	// logs new logs are saved. This is not a big deal since reloading the page
	// would then display the missed logs.
	aliasName, bucket, prefix := request.GetCloudAttributes(project.UID)
	cloudLogs, status, err := helpers.GetLogs(aliasName, bucket, prefix)
	if status == "" {
		// because it will be used as the image path, we display an empty image
		status = "empty"
	}
	if err != nil {
		helpers.RedirectWithErrorFlash("/", fmt.Sprintf("failed to read the "+
			"logs at dedis/%s/logs/%d/ - %s", project.UID, request.Index,
			err.Error()), w, r, store)
		return
	}

	p := &viewData{
		Title:           "List of projects",
		Flash:           flashes,
		Project:         project,
		Request:         request,
		RID:             rid,
		CloudLogs:       cloudLogs,
		CloudLogsStatus: status,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		log.Errorf("Error while executing template: %s\n", err.Error())
	}
}

func projectsRequestShowCloudstreamGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Error("the response writer does not support streaming")
		helpers.RedirectWithErrorFlash("/", "the response writer does not "+
			"support streaming", w, r, store)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "false")

	params := mux.Vars(r)

	pid := params["pid"]
	if pid == "" {
		fmt.Fprintf(w, "data: %s\n\n", "project not found")
		flusher.Flush()
		return
	}

	ridStr := params["rid"]
	if ridStr == "" {
		fmt.Fprintf(w, "data: %s\n\n", "failed to get request id in url")
		flusher.Flush()
		return
	}

	rid, err := strconv.Atoi(ridStr)
	if err != nil {
		fmt.Fprintf(w, "data: %s\n\n", "request id is not an int")
		flusher.Flush()
		return
	}

	project, ok := models.ProjectList[pid]
	if !ok || project == nil {
		fmt.Fprintf(w, "data: %s\n\n", "project not found")
		flusher.Flush()
		return
	}

	if len(project.Requests) <= rid || rid < 0 {
		fmt.Fprintf(w, "data: %s\n\n", "request index is out of bound")
		flusher.Flush()
		return
	}

	request := project.Requests[rid]
	aliasName, bucket, prefix := request.GetCloudAttributes(project.UID)
	cloudNotif, err := helpers.NewCloudNotifier(aliasName, bucket, prefix)
	if err != nil {
		fmt.Fprintf(w, "data: %s\n\n", "failed to get cloud chan: "+err.Error())
		flusher.Flush()
		return
	}

	for {
		select {
		case taskEvent := <-cloudNotif.TaskEventCh:
			fmt.Fprintf(w, taskEvent.JSONStream())
			flusher.Flush()
		default:
			select {
			case <-cloudNotif.Done:
				// We ensure we empty our TaskEventCh before exiting
			streamFlush:
				for {
					select {
					case taskEvent := <-cloudNotif.TaskEventCh:
						fmt.Fprint(w, taskEvent.JSONStream())
						flusher.Flush()
					default:
						break streamFlush
					}
				}
				w.WriteHeader(http.StatusNoContent)
				return
			default:
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func projectsRequestShowStatusStreamGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Error("the response writer does not support streaming")
		helpers.RedirectWithErrorFlash("/", "the response writer does not "+
			"support streaming", w, r, store)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "false")

	params := mux.Vars(r)

	pid := params["pid"]
	if pid == "" {
		fmt.Fprintf(w, "data: %s\n\n", "project not found")
		flusher.Flush()
		return
	}

	ridStr := params["rid"]
	if ridStr == "" {
		fmt.Fprintf(w, "data: %s\n\n", "failed to get request id in url")
		flusher.Flush()
		return
	}

	rid, err := strconv.Atoi(ridStr)
	if err != nil {
		fmt.Fprintf(w, "data: %s\n\n", "request id is not an int")
		flusher.Flush()
		return
	}

	project, ok := models.ProjectList[pid]
	if !ok || project == nil {
		fmt.Fprintf(w, "data: %s\n\n", "project not found")
		flusher.Flush()
		return
	}

	if len(project.Requests) <= rid || rid < 0 {
		fmt.Fprintf(w, "data: %s\n\n", "request index is out of bound")
		flusher.Flush()
		return
	}

	request := project.Requests[rid]

	statusNotifier := request.StatusNotifier

	// If the status notifier is terminated, then no need for a client
	var client *helpers.StatusNotifierSubscriber
	if !statusNotifier.Terminated {
		client = statusNotifier.Subscribe()
	} else {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for {
		select {
		case status := <-client.NotifyStream:
			fmt.Fprintf(w, "data: %s\n\n", status)
			flusher.Flush()
		default:
			select {
			case <-client.Done:
				// We ensure we empty our Taskstream before exiting
			streamFlush:
				for {
					select {
					case status := <-client.NotifyStream:
						fmt.Fprintf(w, "data: %s\n\n", status)
						flusher.Flush()
					default:
						break streamFlush
					}
				}
				log.Info("client done...")
				return
			default:
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func projectsRequestTasksShowStream(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Error("the response writer does not support streaming")
		helpers.RedirectWithErrorFlash("/", "the response writer does not "+
			"support streaming", w, r, store)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "false")

	params := mux.Vars(r)
	pid := params["pid"]
	if pid == "" {
		fmt.Fprintf(w, "data: %s\n\n", "failed to get project id")
		flusher.Flush()
		return
	}

	ridStr := params["rid"]
	if ridStr == "" {
		fmt.Fprintf(w, "data: %s\n\n", "failed to get request id")
		flusher.Flush()
		return
	}

	rid, err := strconv.Atoi(ridStr)
	if err != nil {
		fmt.Fprintf(w, "data: %s\n\n", "request id is not an int")
		flusher.Flush()
		return
	}

	tidStr := params["tid"]
	if tidStr == "" {
		fmt.Fprintf(w, "data: %s\n\n", "failed to get task id")
		flusher.Flush()
		return
	}

	tid, err := strconv.Atoi(tidStr)
	if err != nil {
		fmt.Fprintf(w, "data: %s\n\n", "task id is not an int")
		flusher.Flush()
		return
	}

	project, ok := models.ProjectList[pid]
	if !ok || project == nil {
		fmt.Fprintf(w, "data: %s\n\n", "project not found")
		flusher.Flush()
		return
	}

	if len(project.Requests) <= rid || rid < 0 {
		fmt.Fprintf(w, "data: %s\n\n", "request id is out of bound")
		flusher.Flush()
		return
	}

	request := project.Requests[rid]

	if len(request.Tasks) <= tid || tid < 0 {
		fmt.Fprintf(w, "data: %s\n\n", "task id is out of bound")
		flusher.Flush()
		return
	}

	task := request.Tasks[tid]

	// If the task is not in a working state then there is nothing to return
	// here.
	var client *helpers.Subscriber
	if task.Status == helpers.StatusWorking {
		client = task.Subscribe()
	} else {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var taskEl *helpers.TaskEvent
	// for _, taskEl = range client.PastEvents {
	// 	fmt.Fprint(w, taskEl.JSONStream())
	// 	flusher.Flush()
	// 	time.Sleep(time.Second * 1)
	// }

	for {
		select {
		case taskEl = <-client.TaskStream:
			fmt.Fprint(w, taskEl.JSONStream())
			flusher.Flush()
		default:
			select {
			case <-client.Done:
				// We ensure we empty our TaskStream before exiting
			streamFlush:
				for {
					select {
					case taskEl = <-client.TaskStream:
						fmt.Fprint(w, taskEl.JSONStream())
						flusher.Flush()
					default:
						break streamFlush
					}
				}
				log.Info("client done...")
				return
			default:
			}
		}
		time.Sleep(time.Millisecond * 100)
	}

}

func projectsRequestTasksShowDebugGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	params := mux.Vars(r)
	pid := params["pid"]
	if pid == "" {
		helpers.RedirectWithErrorFlash("/", "project id not found", w, r, store)
		return
	}

	ridStr := params["rid"]
	if ridStr == "" {
		helpers.RedirectWithErrorFlash("/", "request id not found", w, r, store)
		return
	}

	rid, err := strconv.Atoi(ridStr)
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "request id is not an int", w, r, store)
		return
	}

	tidStr := params["tid"]
	if tidStr == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get task id", w, r, store)
		return
	}

	tid, err := strconv.Atoi(tidStr)
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "task id is not an int", w, r, store)
		return
	}

	project, ok := models.ProjectList[pid]
	if !ok || project == nil {
		helpers.RedirectWithErrorFlash("/", "project not found", w, r, store)
		return
	}

	if len(project.Requests) <= rid || rid < 0 {
		helpers.RedirectWithErrorFlash("/", "request id is out of bound", w, r, store)
		return
	}

	request := project.Requests[rid]

	if len(request.Tasks) <= tid || tid < 0 {
		helpers.RedirectWithErrorFlash("/", "task id is out of bound", w, r, store)
		return
	}

	t, err := template.ParseFiles("views/layout.gohtml", "views/projects/requests/tasks/debug.gohtml")
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	type viewData struct {
		Title   string
		Project *models.Project
		Request *models.Request
		Task    *helpers.Task
		Flash   []helpers.Flash
		Tid     int
	}

	flashes, err := helpers.ExtractFlash(w, r, store)
	if err != nil {
		log.Errorf("Failed to get flash: %s\n", err.Error())
	}

	task := request.Tasks[tid]

	p := &viewData{
		Title:   "List of projects",
		Flash:   flashes,
		Project: project,
		Request: request,
		Task:    task,
		Tid:     tid,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		log.Errorf("Error while executing template: %s\n", err.Error())
	}

}

func projectsRequestTasksShowStatusPut(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {
	// Here we assume that "r.ParseForm" has already been called

	params := mux.Vars(r)
	pid := params["pid"]
	if pid == "" {
		helpers.RedirectWithErrorFlash("/", "project id not found", w, r, store)
		return
	}

	ridStr := params["rid"]
	if ridStr == "" {
		helpers.RedirectWithErrorFlash("/", "request id not found", w, r, store)
		return
	}

	rid, err := strconv.Atoi(ridStr)
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "request id is not an int", w, r, store)
		return
	}

	tidStr := params["tid"]
	if tidStr == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get task id", w, r, store)
		return
	}

	tid, err := strconv.Atoi(tidStr)
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "task id is not an int", w, r, store)
		return
	}

	project, ok := models.ProjectList[pid]
	if !ok || project == nil {
		helpers.RedirectWithErrorFlash("/", "project not found", w, r, store)
		return
	}

	if len(project.Requests) <= rid || rid < 0 {
		helpers.RedirectWithErrorFlash("/", "request id is out of bound", w, r, store)
		return
	}

	request := project.Requests[rid]

	if len(request.Tasks) <= tid || tid < 0 {
		helpers.RedirectWithErrorFlash("/", "task id is out of bound", w, r, store)
		return
	}

	task := request.Tasks[tid]

	status := r.PostForm.Get("status")
	if status == "" {
		helpers.RedirectWithErrorFlash("/projects/"+project.UID,
			"status argument not found in form", w, r, store)
		return
	}

	// If the task is currently working but the new status says that it's either
	// a closeOK or closeError, we must then notify the subscribers and set the
	// associated request status accordingly. Warning: this can be dangerous as
	// the request could be closed twice, will would then close the subscribers
	// channels that are already closed.
	if task.Status == helpers.StatusWorking {
		switch status {
		case helpers.StatusFinished:
			task.CloseOK("ds manager (debug)", "manual update of the status", "")
			request.Status = models.RequestStatusDone
			request.StatusNotifier.UpdateStatusAndClose(
				models.RequestStatusDone)
		case helpers.StatusErrored:
			task.CloseError("ds manager (debug)", "manual update of the status", "")
			request.Status = models.RequestStatusErrored
			request.StatusNotifier.UpdateStatusAndClose(
				models.RequestStatusErrored)
		}
	} else {
		task.Status = helpers.StatusTask(status)
	}

	helpers.RedirectWithInfoFlash("/projects/"+project.UID+"/requests/"+ridStr, fmt.Sprintf(
		"Task updated with status '%s'", status), w, r, store)
}
