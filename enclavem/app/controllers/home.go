package controllers

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/dedis/odyssey/enclavem/app/models"
	"github.com/gorilla/sessions"
)

// HomeHandler ...
func HomeHandler(gs sessions.Store, conf *models.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HomeGet(w, r, gs, conf)
		}
	}
}

// HomeGet ...
func HomeGet(w http.ResponseWriter, r *http.Request, gs sessions.Store, conf *models.Config) {
	type viewData struct {
		Title string
		Flash []helpers.Flash
	}

	if r.URL.Path != "/" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	t, err := template.ParseFiles("views/layout.gohtml", "views/home.gohtml")
	if err != nil {
		fmt.Printf("Error with template: %s\n", err.Error())
		helpers.AddFlash(w, r, fmt.Sprintf("<pre>Error with template:\n%s</pre>", err.Error()), gs, helpers.Error)
	}

	flashes, err := helpers.ExtractFlash(w, r, gs)
	if err != nil {
		fmt.Printf("Failed to get flash: %s\n", err.Error())
	}

	p := &viewData{
		Title: "Home",
		Flash: flashes,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		fmt.Printf("Error while executing template: %s\n", err.Error())
	}
}
