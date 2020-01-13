package controllers

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/dedis/odyssey/domanager/app/models"
	xhelpers "github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/gorilla/sessions"
)

// HomeHandler ...
func HomeHandler(store *sessions.CookieStore, conf *models.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HomeGet(w, r, store, conf)
		}
	}
}

// HomeGet ...
func HomeGet(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, conf *models.Config) {
	type viewData struct {
		Title   string
		Flash   []xhelpers.Flash
		Session *models.Session
	}

	if r.URL.Path != "/" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	t, err := template.ParseFiles("views/layout.gohtml", "views/home.gohtml")
	if err != nil {
		fmt.Printf("Error with template: %s\n", err.Error())
		xhelpers.AddFlash(w, r, fmt.Sprintf("<pre>Error with template:\n%s</pre>", err.Error()), store, xhelpers.Error)
	}

	flashes, err := xhelpers.ExtractFlash(w, r, store)
	if err != nil {
		fmt.Printf("Failed to get flash: %s\n", err.Error())
	}

	session, err := models.GetSession(store, r)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to get session: "+
			err.Error(), w, r, store)
		return
	}

	p := &viewData{
		Title:   "Home",
		Flash:   flashes,
		Session: session,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		fmt.Printf("Error while executing template: %s\n", err.Error())
	}
}
