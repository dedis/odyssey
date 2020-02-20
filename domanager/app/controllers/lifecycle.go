package controllers

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"text/template"

	"github.com/dedis/odyssey/catalogc"
	"github.com/dedis/odyssey/domanager/app/models"
	xhelpers "github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/gorilla/sessions"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/protobuf"
	"golang.org/x/xerrors"
)

// ShowLifecycle ...
func ShowLifecycle(store *sessions.CookieStore, conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			lifecycleGet(w, r, store, conf)
		}
	}
}

func lifecycleGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	type viewData struct {
		Title            string
		Flash            []xhelpers.Flash
		Session          *models.Session
		AuditData        catalogc.AuditData
		ShortPID         string
		DataScientistID  string
		EnclaveManagerID string
		EnclaveID        string
	}

	session, err := models.GetSession(store, r)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to get session: "+
			err.Error(), w, r, store)
		return
	}
	if !session.IsLogged() {
		xhelpers.RedirectWithWarningFlash("/", "you need to be logged in to "+
			"access this page", w, r, store)
		return
	}

	piid, ok := r.URL.Query()["piid"]
	if !ok || len(piid[0]) < 1 {
		xhelpers.RedirectWithWarningFlash("/", "'piid' is empty or absent", w, r, store)
		return
	}

	// Check that the piid is well formated
	piidPattern, err := regexp.Compile("^[0-9a-f]{64}$")
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to build piid regex: "+err.Error(), w, r, store)
		return
	}

	if !piidPattern.MatchString(piid[0]) {
		xhelpers.RedirectWithErrorFlash("/", "wrong instance id, the content "+
			"of the project instance id is unexpected:"+piid[0], w, r, store)
		return
	}

	cmd := exec.Command("./catadmin", "-c", conf.ConfigPath, "audit",
		"project", "-i", piid[0], "-bc", session.BcPath)

	log.Info(fmt.Sprintf("command created: %s", cmd.Args))
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf("failed to get the "+
			"lifecycle: %s - Output: %s - Err: %s", err.Error(),
			outb.String(), errb.String()), w, r, store)
		return
	}

	auditData := catalogc.AuditData{}
	err = protobuf.Decode(outb.Bytes(), &auditData)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to decode audit data: "+err.Error(), w, r, store)
	}

	dataScientistID, err := func() (string, error) {
		for _, block := range auditData.Blocks {
			for _, tx := range block.Transactions {
				for _, instr := range tx.Instructions {
					// we assume that only the data scientist signs the spawn request
					if instr.Spawn != nil {
						return instr.SignerIdentities[0].String(), nil
					}
				}
			}
		}
		return "", xerrors.New("spawn instruction not found")
	}()
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to get the datascientist id: "+err.Error(), w, r, store)
	}

	// return an empty string if the invoke:setURL instruction is not found
	enclaveManagerID := func() string {
		for _, block := range auditData.Blocks {
			for _, tx := range block.Transactions {
				for _, instr := range tx.Instructions {
					// Only the enclave manager can call the "setURL" action
					if instr.Invoke != nil && instr.Invoke.Command == "setURL" {
						return instr.SignerIdentities[0].String()
					}
				}
			}
		}
		return ""
	}()

	// return an empty string if the invoke:setEnclavePubKey instruction is not
	// found
	enclaveID := func() string {
		for _, block := range auditData.Blocks {
			for _, tx := range block.Transactions {
				for _, instr := range tx.Instructions {
					// Only the enclave manager can call the "setURL" action
					if instr.Invoke != nil && instr.Invoke.Command == "setEnclavePubKey" {
						pubKey := instr.Invoke.Args.Search("pubKey")
						return string(pubKey)
					}
				}
			}
		}
		return ""
	}()

	t, err := template.New("lifecycle").Funcs(template.FuncMap{
		"toString": func(buf []byte) string {
			return string(buf)
		},
	}).ParseFiles("views/layout.gohtml", "views/lifecycle.gohtml")
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
		Title:            "Lifecycle",
		Flash:            flashes,
		Session:          session,
		AuditData:        auditData,
		ShortPID:         piid[0][:8] + "...",
		DataScientistID:  dataScientistID,
		EnclaveManagerID: enclaveManagerID,
		EnclaveID:        enclaveID,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf(
			"Error while executing template: %s\n", err.Error()), w, r, store)
		return
	}

}
