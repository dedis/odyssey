package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/dedis/odyssey/domanager/app/models"
	xhelpers "github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// ShowtasksIndexHandler ...
func ShowtasksIndexHandler(store sessions.Store, conf *models.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			showtasksGet(w, r, store, conf)
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				xhelpers.RedirectWithErrorFlash(r.URL.String(), "failed to read form", w, r, store)
				return
			}
			switch r.PostFormValue("_method") {
			case "delete":
				showtasksIndexDelete(w, r, store, conf)
			default:
				xhelpers.RedirectWithErrorFlash("/", "only DELETE allowed", w, r, store)
				return
			}
		}
	}
}

// ShowtasksShowHandler ...
func ShowtasksShowHandler(store sessions.Store,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			showtaskShowGet(w, r, store, conf)
		}
	}
}

func showtasksGet(w http.ResponseWriter, r *http.Request, store sessions.Store, conf *models.Config) {

	t, err := template.ParseFiles("views/layout.gohtml", "views/tasks/index.gohtml")
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	type viewData struct {
		Title   string
		Tasks   []xhelpers.TaskI
		Flash   []xhelpers.Flash
		Session *models.Session
	}

	flashes, err := xhelpers.ExtractFlash(w, r, store)
	if err != nil {
		fmt.Printf("Failed to get flash: %s\n", err.Error())
	}

	taskSlice := conf.TaskManager.GetSortedTasks()

	session, err := models.GetSession(store, r)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to get session: "+
			err.Error(), w, r, store)
		return
	}

	p := &viewData{
		Title:   "List of datasets",
		Flash:   flashes,
		Tasks:   taskSlice,
		Session: session,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "error while executing template: "+err.Error(), w, r, store)
		return
	}
}

func showtasksIndexDelete(w http.ResponseWriter, r *http.Request, store sessions.Store, conf *models.Config) {

	conf.TaskManager.DeleteAllTasks()

	xhelpers.RedirectWithInfoFlash("/showtasks", "tasks deleted", w, r, store)
}

func showtaskShowGet(w http.ResponseWriter, r *http.Request, store sessions.Store, conf *models.Config) {

	params := mux.Vars(r)
	indexStr := params["id"]
	if indexStr == "" {
		xhelpers.RedirectWithErrorFlash("/", "failed to get the task index in url", w, r, store)
		return
	}
	index, err := strconv.Atoi(indexStr)

	t, err := template.ParseFiles("views/layout.gohtml", "views/tasks/show.gohtml")
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	type viewData struct {
		Title   string
		Task    xhelpers.TaskI
		Flash   []xhelpers.Flash
		Session *models.Session
	}

	flashes, err := xhelpers.ExtractFlash(w, r, store)
	if err != nil {
		fmt.Printf("Failed to get flash: %s\n", err.Error())
		xhelpers.RedirectWithErrorFlash("/", "Failed to get flash: "+err.Error(), w, r, store)
		return
	}

	if index >= conf.TaskManager.NumTasks() || index < 0 {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf("Index out of bound: "+
			"0 > (index) %d >= len(TaskList) %d", index, conf.TaskManager.NumTasks()),
			w, r, store)
		return
	}
	task := conf.TaskManager.GetTask(index)

	session, err := models.GetSession(store, r)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to get session: "+
			err.Error(), w, r, store)
		return
	}

	p := &viewData{
		Title:   "Request " + task.GetData().ID + " with index " + string(task.GetData().Index),
		Flash:   flashes,
		Task:    task,
		Session: session,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "rror while executing template: "+err.Error(), w, r, store)
		return
	}
}
