package controllers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path"
	"path/filepath"
	"text/template"

	"github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/gorilla/sessions"
)

// AuthorizeHandler ...
func AuthorizeHandler(gs *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			authorizeGet(w, r, gs)
		case http.MethodPost:
			authorizePost(w, r, gs)
		}
	}
}

func authorizeGet(w http.ResponseWriter, r *http.Request, gs *sessions.CookieStore) {
	type viewData struct {
		Title string
		Flash []helpers.Flash
	}
	// cl := calypso.NewClient(nil)

	t, err := template.ParseFiles("views/layout.gohtml", "views/authorize.gohtml")
	if err != nil {
		fmt.Printf("Error with template: %s\n", err.Error())
	}

	flashes, err := helpers.ExtractFlash(w, r, gs)
	if err != nil {
		fmt.Printf("Failed to get flash: %s\n", err.Error())
	}
	p := &viewData{
		Title: "Authorize",
		Flash: flashes,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		fmt.Printf("Error while executing template: %s\n", err.Error())
	}
}

func authorizePost(w http.ResponseWriter, r *http.Request, gs *sessions.CookieStore) {
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		msg := fmt.Sprintf("Error Retrieving the File: %s\n", err.Error())
		helpers.AddFlash(w, r, msg, gs, helpers.Error)
		authorizeGet(w, r, gs)
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
		helpers.AddFlash(w, r, err.Error(), gs, helpers.Error)
		authorizeGet(w, r, gs)
		return
	}
	newpath, err := filepath.Abs(path.Join("tmp", handler.Filename))
	if err != nil {
		helpers.AddFlash(w, r, err.Error(), gs, helpers.Error)
		authorizeGet(w, r, gs)
		return
	}

	err = ioutil.WriteFile(newpath, fileBytes, 0664)
	if err != nil {
		helpers.AddFlash(w, r, err.Error(), gs, helpers.Error)
		authorizeGet(w, r, gs)
		return
	}
	// return that we have successfully uploaded our file!
	// fmt.Fprintf(w, "Successfully Uploaded File to %s\n", newpath)
	helpers.AddFlash(w, r, fmt.Sprintf("File added to %s", newpath), gs, helpers.Info)

	cmd := exec.Command("csadmin", "authorize", newpath, r.FormValue("bid"))
	if err != nil {
		helpers.AddFlash(w, r, fmt.Sprintf("failed to exec command: %s", err.Error()), gs, helpers.Error)
		authorizeGet(w, r, gs)
		return
	}

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		helpers.AddFlash(w, r, fmt.Sprintf("<pre>failed to run command: %s\nOutput: %s\nErr: %s</pre>", err.Error(), outb.String(), errb.String()), gs, helpers.Error)
		authorizeGet(w, r, gs)
		return
	}

	// cfg, err := app.LoadCothority(newpath)
	// if err != nil {
	// 	helpers.AddFlash(w, r, "failed to load cothority: "+err.Error(), gs, helpers.Error)
	// 	authorizeGet(w, r, gs)
	// 	return
	// }
	// si, err := cfg.GetServerIdentity()
	// if err != nil {
	// 	helpers.AddFlash(w, r, "failed to get server identity: "+err.Error(), gs, helpers.Error)
	// 	authorizeGet(w, r, gs)
	// 	return
	// }

	// bc, err := hex.DecodeString(r.FormValue("bid"))
	// if err != nil {
	// 	helpers.AddFlash(w, r, "failed to decode string: "+err.Error(), gs, helpers.Error)
	// 	authorizeGet(w, r, gs)
	// 	return
	// }
	// log.Infof("Contacting %s to authorize byzcoin %x", si.Address, bc)
	// cl := calypso.NewClient(nil)

	// err = cl.Authorize(si, bc)
	// if err != nil {
	// 	helpers.AddFlash(w, r, "failed to authorize: "+err.Error(), gs, helpers.Error)
	// 	authorizeGet(w, r, gs)
	// 	return
	// }
	helpers.AddFlash(w, r, "Added ByzCoinID to the list of authorized IDs in the server", gs, helpers.Info)

	authorizeGet(w, r, gs)
}
