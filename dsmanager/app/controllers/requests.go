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
func RequestsIndexHandler(gs *sessions.CookieStore, conf *models.Config, db *bolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			requestsGet(w, r, gs, conf)
		}
	}
}

// RequestsShowHandler ...
func RequestsShowHandler(gs *sessions.CookieStore, conf *models.Config, db *bolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			requestsSwhoGet(w, r, gs, conf)
		}
	}
}

func requestsGet(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, conf *models.Config) {

	t, err := template.ParseFiles("views/layout.gohtml", "views/requests/index.gohtml")
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	type viewData struct {
		Title    string
		Requests []*helpers.Task
		Flash    []helpers.Flash
	}

	flashes, err := helpers.ExtractFlash(w, r, store)
	if err != nil {
		fmt.Printf("Failed to get flash: %s\n", err.Error())
	}

	taskSlice := []*helpers.Task{}
	for _, value := range helpers.TaskList {
		taskSlice = append(taskSlice, value)
	}
	p := &viewData{
		Title:    "List of datasets",
		Flash:    flashes,
		Requests: taskSlice,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "rror while executing template: "+err.Error(), w, r, store)
		return
	}
}

func requestsSwhoGet(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, conf *models.Config) {

	params := mux.Vars(r)
	indexStr := params["id"]
	if indexStr == "" {
		helpers.RedirectWithErrorFlash("/", "failed to get the request id in url", w, r, store)
		return
	}
	index, err := strconv.Atoi(indexStr)

	t, err := template.ParseFiles("views/layout.gohtml", "views/requests/show.gohtml")
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	type viewData struct {
		Title     string
		Request   *helpers.Task
		StatusImg string
		Flash     []helpers.Flash
	}

	flashes, err := helpers.ExtractFlash(w, r, store)
	if err != nil {
		fmt.Printf("Failed to get flash: %s\n", err.Error())
		helpers.RedirectWithErrorFlash("/", "Failed to get flash: "+err.Error(), w, r, store)
		return
	}

	if index >= len(helpers.TaskList) || index < 0 {
		helpers.RedirectWithErrorFlash("/", fmt.Sprintf("Index out of bound: "+
			"0 > (index) %d >= len(TaskList) %d", index, len(helpers.TaskList)),
			w, r, store)
		return
	}
	task := helpers.TaskList[index]

	p := &viewData{
		Title:     "Request " + task.ID + " with index " + string(task.Index),
		Flash:     flashes,
		StatusImg: helpers.StatusImage(task.Status),
		Request:   task,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "rror while executing template: "+err.Error(), w, r, store)
		return
	}
}
