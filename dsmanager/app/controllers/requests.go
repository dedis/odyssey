package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/dedis/odyssey/dsmanager/app/models"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	bolt "go.etcd.io/bbolt"
)

// RequestsIndexHandler ...
func RequestsIndexHandler(gs sessions.Store, conf *models.Config, db *bolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			requestsGet(w, r, gs, conf)
		}
	}
}

// RequestsShowHandler ...
func RequestsShowHandler(gs sessions.Store, conf *models.Config, db *bolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			requestsSwhoGet(w, r, gs, conf)
		}
	}
}

func requestsGet(w http.ResponseWriter, r *http.Request,
	store sessions.Store, conf *models.Config) {

	t, err := template.ParseFiles("views/layout.gohtml", "views/requests/index.gohtml")
	if err != nil {
		helpers.RedirectWithErrorFlash("/",
			"Error with template: "+err.Error(), w, r, store)
		return
	}

	type viewData struct {
		Title string
		Tasks []helpers.TaskI
		Flash []helpers.Flash
	}

	flashes, err := helpers.ExtractFlash(w, r, store)
	if err != nil {
		fmt.Printf("Failed to get flash: %s\n", err.Error())
	}

	p := &viewData{
		Title: "List of datasets",
		Flash: flashes,
		Tasks: conf.TaskManager.GetSortedTasks(),
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		helpers.RedirectWithErrorFlash("/",
			"error while executing template: "+err.Error(), w, r, store)
		return
	}
}

func requestsSwhoGet(w http.ResponseWriter, r *http.Request,
	store sessions.Store, conf *models.Config) {

	params := mux.Vars(r)
	indexStr := params["id"]
	if indexStr == "" {
		helpers.RedirectWithErrorFlash("/",
			"failed to get the request id in url", w, r, store)
		return
	}
	index, err := strconv.Atoi(indexStr)
	if indexStr == "" {
		helpers.RedirectWithErrorFlash("/",
			"failed to convert indexStr: "+err.Error(), w, r, store)
		return
	}

	t, err := template.ParseFiles("views/layout.gohtml",
		"views/requests/show.gohtml")
	if err != nil {
		helpers.RedirectWithErrorFlash("/",
			"Error with template: "+err.Error(), w, r, store)
		return
	}

	type viewData struct {
		Title string
		Task  helpers.TaskI
		Flash []helpers.Flash
	}

	flashes, err := helpers.ExtractFlash(w, r, store)
	if err != nil {
		fmt.Printf("Failed to get flash: %s\n", err.Error())
		helpers.RedirectWithErrorFlash("/",
			"Failed to get flash: "+err.Error(), w, r, store)
		return
	}

	if index >= conf.TaskManager.NumTasks() || index < 0 {
		helpers.RedirectWithErrorFlash("/", fmt.Sprintf("Index out of bound: "+
			"0 > (index) %d >= len(TaskList) %d", index,
			conf.TaskManager.NumTasks()), w, r, store)
		return
	}
	task := conf.TaskManager.GetTask(index)

	p := &viewData{
		Title: "Request " + task.GetData().ID + " with index " +
			string(task.GetData().Index),
		Flash: flashes,
		Task:  task,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		helpers.RedirectWithErrorFlash("/",
			"error while executing template: "+err.Error(), w, r, store)
		return
	}
}
