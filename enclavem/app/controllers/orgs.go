package controllers

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dedis/odyssey/enclavem/app/helpers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"go.dedis.ch/onet/v3/log"
)

// OrgsIndexHandler points to:
// GET /Orgs
func OrgsIndexHandler(gs *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			OrgsIndexGet(w, r, gs)
		}
	}
}

// OrgsShowHandler points to:
// GET /Orgs/{id}
func OrgsShowHandler(gs *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			OrgsShowGet(w, r, gs)
		}
	}
}

// OrgsIndexGet return a json representation of a list of orgs
func OrgsIndexGet(w http.ResponseWriter, r *http.Request, gs *sessions.CookieStore) {

	token, err := helpers.GetToken(w)
	if err != nil {
		helpers.SendRequestError(err, w)
		return
	}

	type OrgList struct {
		XMLName xml.Name `xml:"OrgList"`
		Text    string   `xml:",chardata"`
		Xmlns   string   `xml:"xmlns,attr"`
		Ovf     string   `xml:"ovf,attr"`
		Vssd    string   `xml:"vssd,attr"`
		Common  string   `xml:"common,attr"`
		Rasd    string   `xml:"rasd,attr"`
		Vmw     string   `xml:"vmw,attr"`
		Vmext   string   `xml:"vmext,attr"`
		Ovfenv  string   `xml:"ovfenv,attr"`
		Ns9     string   `xml:"ns9,attr"`
		Href    string   `xml:"href,attr"`
		Type    string   `xml:"type,attr"`
		Org     struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
			Name string `xml:"name,attr"`
			Type string `xml:"type,attr"`
		} `xml:"Org"`
	}

	url := fmt.Sprintf("https://%s/api/org", helpers.VcdHost)
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

	var orgs OrgList
	err = xml.Unmarshal(bodyBuf, &orgs)
	if err != nil {
		helpers.SendRequestError(errors.New("failed to unmarshal body: "+err.Error()), w)
		return
	}

	js, err := json.MarshalIndent(orgs, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// OrgsShowGet gets a single org and sends a json representation of it
func OrgsShowGet(w http.ResponseWriter, r *http.Request, gs *sessions.CookieStore) {
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

	type Org struct {
		XMLName xml.Name `xml:"Org"`
		Text    string   `xml:",chardata"`
		Xmlns   string   `xml:"xmlns,attr"`
		Ovf     string   `xml:"ovf,attr"`
		Vssd    string   `xml:"vssd,attr"`
		Common  string   `xml:"common,attr"`
		Rasd    string   `xml:"rasd,attr"`
		Vmw     string   `xml:"vmw,attr"`
		Vmext   string   `xml:"vmext,attr"`
		Ovfenv  string   `xml:"ovfenv,attr"`
		Ns9     string   `xml:"ns9,attr"`
		Name    string   `xml:"name,attr"`
		ID      string   `xml:"id,attr"`
		Href    string   `xml:"href,attr"`
		Type    string   `xml:"type,attr"`
		Link    []struct {
			Text string `xml:",chardata"`
			Rel  string `xml:"rel,attr"`
			Href string `xml:"href,attr"`
			Name string `xml:"name,attr"`
			Type string `xml:"type,attr"`
		} `xml:"Link"`
		Description string `xml:"Description"`
		FullName    string `xml:"FullName"`
	}

	url := fmt.Sprintf("https://%s/api/org/%s", helpers.VcdHost, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.SendRequestError(errors.New("failed to build request:"+err.Error()), w)
		return
	}
	req.Header.Set("x-vcloud-authorization", token)
	req.Header.Set("Accept", "application/*;version=27.0")

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

	var org Org
	err = xml.Unmarshal(bodyBuf, &org)
	if err != nil {
		helpers.SendRequestError(errors.New("failed to unmarshal body: "+err.Error()), w)
		return
	}

	js, err := json.MarshalIndent(org, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
