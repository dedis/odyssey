package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"text/template"

	"github.com/dedis/odyssey/domanager/app/models"
	xhelpers "github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/gorilla/sessions"
)

// SessionHandler ...
func SessionHandler(store *sessions.CookieStore, conf *models.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			sessionGet(w, r, store, conf)
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				xhelpers.RedirectWithErrorFlash(r.URL.String(), "failed to read form", w, r, store)
				return
			}
			switch r.PostFormValue("_method") {
			case "delete":
				sessionDelete(w, r, store, conf)
			default:
				sessionPost(w, r, store, conf)
			}
		}
	}
}

// SessionProfileHandler ...
func SessionProfileHandler(store *sessions.CookieStore, conf *models.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			profileGet(w, r, store, conf)
		}
	}
}

func sessionGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	type viewData struct {
		Title   string
		Flash   []xhelpers.Flash
		Session *models.Session
	}

	session, err := models.GetSession(store, r)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to get session: "+
			err.Error(), w, r, store)
		return
	}

	t, err := template.ParseFiles("views/layout.gohtml", "views/sessions/signin.gohtml")
	if err != nil {
		fmt.Printf("Error with template: %s\n", err.Error())
		xhelpers.RedirectWithErrorFlash("/",
			fmt.Sprintf("<pre>Error with template:\n%s</pre>", err.Error()), w, r, store)
		return
	}

	flashes, err := xhelpers.ExtractFlash(w, r, store)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to extract flash", w, r, store)
		return
	}

	p := &viewData{
		Title:   "Signin",
		Flash:   flashes,
		Session: session,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf("Error while executing template: %s\n", err.Error()), w, r, store)
		return
	}
}

func sessionPost(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		msg := fmt.Sprintf("Error Retrieving the File: %s\n", err.Error())
		xhelpers.RedirectWithErrorFlash("/", msg, w, r, store)
		return
	}
	defer file.Close()
	// fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	// fmt.Printf("File Size: %+v\n", handler.Size)
	// fmt.Printf("MIME Header: %+v\n", handler.Header)

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to read the file: "+
			err.Error(), w, r, store)
		return
	}
	newpath, err := filepath.Abs(path.Join("uploads", handler.Filename))
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to get the path: "+
			err.Error(), w, r, store)
		return
	}

	err = ioutil.WriteFile(newpath, fileBytes, 0664)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to write file: "+err.Error(), w, r, store)
		return
	}
	// return that we have successfully uploaded our file!
	// fmt.Fprintf(w, "Successfully Uploaded File to %s\n", newpath)
	xhelpers.AddFlash(w, r, fmt.Sprintf("File added to %s", newpath), store, xhelpers.Info)

	err = models.LogIn(newpath, store, r, w)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to log in: "+err.Error(), w, r, store)
	}

	xhelpers.RedirectWithInfoFlash("/", "Identity file uploaded and set!", w, r, store)
}

func sessionDelete(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	session, err := models.GetSession(store, r)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to get session: "+
			err.Error(), w, r, store)
		return
	}

	if !session.IsLogged() {
		xhelpers.RedirectWithWarningFlash("/", "your are not logged", w, r, store)
		return
	}

	err = session.Destroy(store, r, w)

	xhelpers.RedirectWithInfoFlash("/", "Session deleted", w, r, store)
}

func profileGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	type viewData struct {
		Title   string
		Flash   []xhelpers.Flash
		Session *models.Session
	}

	session, err := models.GetSession(store, r)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to get session: "+
			err.Error(), w, r, store)
		return
	}

	if !session.IsLogged() {
		xhelpers.RedirectWithWarningFlash("/", "your are not logged", w, r, store)
		return
	}

	t, err := template.ParseFiles("views/layout.gohtml", "views/sessions/profile.gohtml")
	if err != nil {
		fmt.Printf("Error with template: %s\n", err.Error())
		xhelpers.RedirectWithErrorFlash("/",
			fmt.Sprintf("<pre>Error with template:\n%s</pre>", err.Error()), w, r, store)
		return
	}

	flashes, err := xhelpers.ExtractFlash(w, r, store)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to extract flash", w, r, store)
		return
	}

	p := &viewData{
		Title:   "Profile",
		Flash:   flashes,
		Session: session,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf("Error while executing template: %s\n", err.Error()), w, r, store)
		return
	}
}
