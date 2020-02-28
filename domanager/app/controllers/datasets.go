package controllers

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/dedis/odyssey/catalogc"
	"github.com/dedis/odyssey/domanager/app/models"
	xhelpers "github.com/dedis/odyssey/dsmanager/app/helpers"
	enclavemodels "github.com/dedis/odyssey/enclavem/app/models"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/minio/minio-go/v6"
	"go.dedis.ch/onet/v3/log"
)

// DatasetIndexHandler ...
func DatasetIndexHandler(store *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			datasetsGet(w, r, store, conf)
		case http.MethodPost:
			datasetsPost(w, r, store, conf)
		}
	}
}

// DatasetNewHandler ...
func DatasetNewHandler(store *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			datasetsNew(w, r, store, conf)
		}
	}
}

// DatasetShowHandler ...
func DatasetShowHandler(store *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			datasetsShow(w, r, store, conf)
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				xhelpers.RedirectWithErrorFlash(r.URL.String(),
					"failed to read form", w, r, store)
				return
			}
			switch r.PostFormValue("_method") {
			case "put":
				datasetsShowPut(w, r, store, conf)
			case "delete":
				datasetsShowDelete(w, r, store, conf)
			default:
				xhelpers.RedirectWithErrorFlash(r.URL.String(),
					"only PUT allowed", w, r, store)
			}
		}
	}
}

// DatasetShowAttributesHandler ...
func DatasetShowAttributesHandler(store *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				xhelpers.RedirectWithErrorFlash(r.URL.String(),
					"failed to read form", w, r, store)
				return
			}
			switch r.PostFormValue("_method") {
			case "put":
				datasetsShowAttributesPut(w, r, store, conf)
			default:
				xhelpers.RedirectWithErrorFlash(r.URL.String(),
					"only PUT allowed", w, r, store)
			}
		}
	}
}

// DatasetShowArchiveHandler ...
func DatasetShowArchiveHandler(store *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				xhelpers.RedirectWithErrorFlash(r.URL.String(),
					"failed to read form", w, r, store)
				return
			}
			switch r.PostFormValue("_method") {
			case "put":
				datasetsShowArchivePut(w, r, store, conf)
			default:
				xhelpers.RedirectWithErrorFlash(r.URL.String(),
					"only PUT allowed", w, r, store)
			}
		}
	}
}

// DatasetShowAuditHandler ...
func DatasetShowAuditHandler(store *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			datasetsShowAuditGet(w, r, store, conf)
		}
	}
}

// DatasetShowDebugHandler ...
func DatasetShowDebugHandler(store *sessions.CookieStore,
	conf *models.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			datasetsShowDebugGet(w, r, store, conf)
		}
	}
}

func datasetsGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	type viewData struct {
		Title    string
		Flash    []xhelpers.Flash
		Session  *models.Session
		Datasets []*catalogc.Dataset
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

	identityStr := session.Cfg.AdminIdentity.String()
	cmd := exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
		"catalog", "getDatasets", "-i", conf.CatalogID, "-bc", session.BcPath,
		"-identityStr", identityStr, "--toJson")

	log.Info(fmt.Sprintf("command created: %s", cmd.Args))
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf("failed to get the "+
			"list of datasets: %s - Output: %s - Err: %s", err.Error(),
			outb.String(), errb.String()), w, r, store)
		return
	}
	datasets := make([]*catalogc.Dataset, 0)
	err = json.Unmarshal(outb.Bytes(), &datasets)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to unmarshal result: "+
			err.Error(), w, r, store)
	}

	if len(datasets) == 1 && datasets[0] == nil {
		datasets = []*catalogc.Dataset{}
	}

	t, err := template.ParseFiles("views/layout.gohtml", "views/datasets/index.gohtml")
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
		Title:    "Your datasets",
		Flash:    flashes,
		Session:  session,
		Datasets: datasets,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf(
			"Error while executing template: %s\n", err.Error()), w, r, store)
		return
	}
}

var darcR = regexp.MustCompile("^darc:[0-9a-f]{64}$")
var baseMatch = regexp.MustCompile(`BaseID: (darc:[0-9a-f]{64})`)
var datasetR = regexp.MustCompile("^[0-9a-f]{64}$")
var darcMatch = regexp.MustCompile(`DarcID: ([0-9a-f]{64})`)
var readAttrRule = regexp.MustCompile(
	// we are interested only by the calypso read rule (note the space)
	`spawn:calypsoRead - ` +
		// we catpure only if the "attr:allowed" is preceded by some spaces
		// (if it is in the middle of the rule), or by " (if it is at the
		// begining)
		`.*["\s]+` +
		// this is the uniq identifier of our attribue
		`attr:allowed:` +
		// we capture everything that is not a white space (if the rule is
		// in the middle) or a " (if the rule is at the end)
		`([^\s"]*)` +
		// then there might be some other rules after that
		`.*`)
var readRule = regexp.MustCompile(`spawn:calypsoRead - "(.*)"`)

// In this regex we capture `( attr:allowed: ... & ... attr:must_have: ... )`
// \b is a word boundary, [^\s] means any char except whitespace.
var attrRule = regexp.MustCompile(`.*\b*(\(\s*attr:allowed:[^\s]*\s*&\s*attr:must_have:[^\s]*\s*\))`)

func datasetsPost(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

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

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to parse form: "+
			err.Error(), w, r, store)
		return
	}

	missing := ""
	title := r.PostFormValue("title")
	if title == "" {
		missing += "- 'title' is empty "
	}
	description := r.PostFormValue("description")
	if description == "" {
		missing += "- 'description' is empty "
	}
	if missing != "" {
		xhelpers.RedirectWithErrorFlash("/datasets/new",
			"Error, some fields were empty: "+missing, w, r, store)
		return
	}

	// creating the task
	tef := xhelpers.NewTaskEventFactory("DO Manager")
	task := xhelpers.NewTask(fmt.Sprintf("Upload of the dataset '%s'", title))
	task.GoBackLink = `<p><a class="pure-button" href="/datasets/new">🔙 Back to the upload of a dataset</a></p>`
	task.AddInfof(tef.Source, "starting the upload process", "got this POST form: %v", r.PostForm)

	go func() {
		var newDarcID string

		if conf.Standalone {

			// Create a darc to control access to the calypsoWriteInstance. For
			// the moment, the only rule on it allows evolution by the admin
			// Darc from the current bc.cfg file.

			desc := fmt.Sprintf("Darc generated by the domanager on %s",
				time.Now().Format("Mon Jan 2 15:04:05 2006"))
			cmd := exec.Command("./bcadmin", "darc", "add",
				"--id", session.Cfg.AdminIdentity.String(),
				"--desc", desc, "-bc", session.BcPath, "--shortPrint",
				"--unrestricted")
			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			err = cmd.Run()
			if err != nil {
				task.CloseError(tef.Source, "create new darc", fmt.Sprintf(
					"failed to create a new DARC: %s - Output: %s - Err: %s",
					err.Error(), outb.String(), errb.String()))
				return
			}

			output := outb.String()
			output = strings.Trim(output, "\n ")
			outputSplit := strings.Split(output, "\n")
			// outputSplit should contain the darcID as the first element and
			// the identity as the second
			if len(outputSplit) != 2 {
				task.CloseError(tef.Source, "parse new darc", fmt.Sprintf(
					"expected the output split to have 2 elements, "+
						"but foud %d: %v", len(outputSplit), outputSplit))
				return
			}

			newDarcID = outputSplit[0]

			// lets check the content of the darc
			ok := darcR.MatchString(newDarcID)
			if !ok {
				task.CloseError(tef.Source, "regexp parse new darc",
					fmt.Sprintf("regex match failed: '%s' does not match the "+
						"darc regex", newDarcID))
				return
			}

			// spawn:calypsoWrite
			cmd = new(exec.Cmd)
			cmd = exec.Command("./bcadmin", "-c", conf.ConfigPath, "darc",
				"rule", "--rule", "spawn:calypsoWrite",
				"--id", session.Cfg.AdminIdentity.String(),
				"-bc", session.BcPath, "--darc", newDarcID)
			outb.Reset()
			errb.Reset()
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			err = cmd.Run()
			if err != nil {
				task.CloseError(tef.Source, "spawn:calypsoRead", fmt.Sprintf(
					"failed to add rule 'spawn:calypsoRead' on the new "+
						"DARC: %s - Output: %s - Err: %s", err.Error(),
					outb.String(), errb.String()))
				return
			}

			// spawn:calypsoRead
			cmd = new(exec.Cmd)
			cmd = exec.Command("./bcadmin", "-c", conf.ConfigPath, "darc",
				"rule", "--rule", "spawn:calypsoRead",
				"--id", session.Cfg.AdminIdentity.String(),
				"-bc", session.BcPath, "--darc", newDarcID)
			outb.Reset()
			errb.Reset()
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			err = cmd.Run()
			if err != nil {
				task.CloseError(tef.Source, "spawn:calypsoRead", fmt.Sprintf(
					"failed to add rule 'spawn:calypsoRead' on the new "+
						"DARC: %s - Output: %s - Err: %s", err.Error(),
					outb.String(), errb.String()))
				return
			}

			task.AddInfof(tef.Source, "Darc created in standalone mode",
				"DarcID: %v", newDarcID)
		} else {
			// Getting a DARC to spawn a new calypso write instance from the
			// enclave manager. Note that there is no identity check from the
			// enclave manager yet, so anyone can request a DARC
			formData := url.Values{
				"darcID": {session.Cfg.AdminDarc.GetIdentityString()},
			}
			task.AddInfof(tef.Source, "getting a new DARC from the enclave manager",
				"calling POST http://localhost:5000/darcs with %v", formData)
			resp, err := http.PostForm("http://localhost:5000/darcs", formData)
			if err != nil {
				task.CloseError(tef.Source, "failed to create PostForm", err.Error())
				return
			}
			defer resp.Body.Close()

			bodyBuff, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				task.CloseError(tef.Source, "failed to read body response", err.Error())
				return
			}

			type Response struct {
				DarcID string `json:"darcID"`
			}

			darcPostResponse := &enclavemodels.DarcPostResponse{}
			err = json.Unmarshal(bodyBuff, darcPostResponse)
			if err != nil {
				task.CloseError(tef.Source, "failed to unmarshal the json response",
					err.Error())
				return
			}
			newDarcID = darcPostResponse.DarcID
		}

		task.AddInfo(tef.Source, "getting the dataset file",
			"retrieving the 'dataset-file' post argument")
		file, handler, err := r.FormFile("dataset-file")
		if err != nil {
			task.CloseError(tef.Source, "Error Retrieving the File", err.Error())
			return
		}
		defer file.Close()

		// Generating the symetric key and the initialization value. We need 16
		// bytes for the key and 12 bytes for the initialization value
		task.AddInfo(tef.Source, "genering a symmetric key",
			"genering a 16 bytes symmetric key")
		key := make([]byte, 16)
		_, err = rand.Read(key)
		if err != nil {
			task.CloseError(tef.Source, "failed to generate key", err.Error())
			return
		}
		task.AddInfo(tef.Source, "genering an nonce", "genering a 12 bytes nonce")
		iv := make([]byte, 12)
		_, err = rand.Read(iv)
		if err != nil {
			task.CloseError(tef.Source, "failed to generate init val", err.Error())
			return
		}

		keyHex := hex.EncodeToString(key)
		ivHex := hex.EncodeToString(iv)

		// Encrypting the dataset
		task.AddInfo(tef.Source, "encrypting the dataset",
			"encrypting with AES using the Galois Counter Mode")
		block, err := aes.NewCipher(key)
		if err != nil {
			task.CloseError(tef.Source, "failed to create new cipher", err.Error())
		}

		aesgcm, err := cipher.NewGCM(block)
		if err != nil {
			task.CloseError(tef.Source, "failed to create new cipher", err.Error())
		}

		fileBuf, err := ioutil.ReadAll(file)
		if err != nil {
			task.CloseError(tef.Source, "failed to readAll the file", err.Error())
			return
		}
		ciphertext := aesgcm.Seal(nil, iv, fileBuf, nil)
		ciphertextReader := bytes.NewReader(ciphertext)

		// Uploading the dataset on the cloud

		extension := filepath.Ext(handler.Filename)
		newFileName := fmt.Sprintf("%s_%s_%s%s.aes", session.GetIdentity(),
			time.Now().Format("2006_01_02_030405"), url.QueryEscape(title), extension)
		cloudURL := fmt.Sprintf("dedis/datasets/%s", newFileName)

		task.AddInfof(tef.Source, "uploading the encrypted dataset on the cloud",
			"saving the encrypted dataset at %s", cloudURL)
		minioClient, err := xhelpers.GetMinioClient()
		if err != nil {
			task.CloseError(tef.Source, "failed to get the minion client", err.Error())
			return
		}
		minioClient.PutObject("datasets", newFileName, ciphertextReader, -1,
			minio.PutObjectOptions{})
		if err != nil {
			task.CloseError(tef.Source, "failed to upload the dataset on the cloud",
				err.Error())
		}

		// Computing the SHA2 of the unencrypted dataset

		task.AddInfo(tef.Source, "computing the SHA2",
			"using the unencrypted file to compute the SHA2")
		h := sha256.New()
		h.Write(fileBuf)
		sha2 := hex.EncodeToString(h.Sum(nil))

		// Creating the calypso write. We need to store the cloud URL because
		// the enclave will get it by parsing the extra data of the write
		// instance with `perl -n -e '/"CloudURL": "(.*?)",/ && print $1'`
		task.AddInfo(tef.Source, "creating a Calypso write", "using csadmin "+
			"to create the calypso write that contains the symetric key and the nonce")
		identityStr := session.Cfg.AdminIdentity.String()
		cmd := exec.Command("./csadmin", "-c", conf.ConfigPath, "contract",
			"write", "spawn", "--darc", newDarcID, "--sign", identityStr, "--bc",
			session.BcPath, "--instid", conf.LtsID, "--secret", keyHex+ivHex,
			"--key", conf.LtsKey, "--extraData", "\"CloudURL\": \""+cloudURL+
				"\", \"IdentityStr\": \""+identityStr+"\"")

		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "csadmin failed", fmt.Sprintf("failed "+
				"to spawn a write instance: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}

		output := outb.String()
		log.Info("Here is the output of the spawn: ", output)
		// We know the csadmin command will output the instID at the second line
		outputSplit := strings.Split(output, "\n")
		if len(outputSplit) < 2 {
			task.CloseError(tef.Source, "got a wrong output", fmt.Sprintf(
				"Got unexpected output split: %s", outputSplit))
			return
		}

		task.AddInfo(tef.Source, "getting the write instance ID",
			"parsing the output of csadmin to extract the write instance ID")
		writeInstID := outputSplit[1]
		ok := datasetR.MatchString(writeInstID)
		if !ok {
			log.Info("got a wong project ID: " + writeInstID)
			task.CloseError(tef.Source, "got a wong write ID", writeInstID)
			return
		}

		log.Info("got this write instance id: " + writeInstID)

		log.Info("now trying to update the catalog")

		cmd = new(exec.Cmd)
		cmd = exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
			"catalog", "invoke", "addDataset", "--sign",
			session.Cfg.AdminIdentity.String(), "--bc", session.BcPath,
			"--instid", conf.CatalogID, "--identityStr", identityStr,
			"--calypsoWriteID", writeInstID, "--title", title, "--description",
			description, "--cloudURL", cloudURL, "--sha2", sha2)
		task.AddInfof(tef.Source, "updating the catalog",
			"using the following command: %v", cmd.Args)

		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		outb.Reset()
		errb.Reset()
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "update catalog failed", fmt.Sprintf(
				"failed to update catalog: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}

		task.CloseOK(tef.Source, "dataset created", outb.String())
	}()

	xhelpers.RedirectWithInfoFlash(fmt.Sprintf("/showtasks/%d", task.Index),
		fmt.Sprintf("Task to create dataset with index %d created", task.Index), w, r, store)
}

func datasetsNew(w http.ResponseWriter, r *http.Request,
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
		xhelpers.RedirectWithWarningFlash("/", "you need to be logged in to "+
			"access this page", w, r, store)
		return
	}

	t, err := template.ParseFiles("views/layout.gohtml",
		"views/datasets/new.gohtml")
	if err != nil {
		fmt.Printf("Error with template: %s\n", err.Error())
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf("<pre>Error with "+
			"template:\n%s</pre>", err.Error()), w, r, store)
		return
	}

	flashes, err := xhelpers.ExtractFlash(w, r, store)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to extract flash",
			w, r, store)
		return
	}

	p := &viewData{
		Title:   "Your datasets",
		Flash:   flashes,
		Session: session,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf("Error while "+
			"executing template: %s\n", err.Error()), w, r, store)
		return
	}
}

func datasetsShow(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		xhelpers.RedirectWithErrorFlash("/", "failed to get the dataset id "+
			"in url", w, r, store)
		return
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

	identityStr := session.Cfg.AdminIdentity.String()
	cmd := exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
		"catalog", "getDatasets", "-i", conf.CatalogID, "-bc", session.BcPath,
		"-identityStr", identityStr, "--toJson")

	log.Info(fmt.Sprintf("command created: %s", cmd.Args))
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf("failed to get the "+
			"list of datasets: %s - Output: %s - Err: %s", err.Error(),
			outb.String(), errb.String()), w, r, store)
		return
	}
	datasets := make([]*catalogc.Dataset, 0)
	err = json.Unmarshal(outb.Bytes(), &datasets)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to unmarshal result: "+
			err.Error(), w, r, store)
		return
	}

	if len(datasets) == 1 && datasets[0] == nil {
		datasets = []*catalogc.Dataset{}
	}

	var dataset *catalogc.Dataset
	for _, d := range datasets {
		if d.CalypsoWriteID == id {
			dataset = d
			break
		}
	}

	if dataset == nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf(
			"dataset with id '%s' not found for identityStr '%s'",
			id, identityStr), w, r, store)
		return
	}

	// Now let's find the DARC associated with this write instance. For this we
	// first need the "bcadmin instance get" command, then the "bcadmin darc
	// show"

	log.Info("getting the darc ID")
	cmd = new(exec.Cmd)
	cmd = exec.Command("./bcadmin", "instance", "get", "-i",
		dataset.CalypsoWriteID, "--bc", session.BcPath)

	log.Info(fmt.Sprintf("command created: %s", cmd.Args))
	outb.Reset()
	errb.Reset()
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf("failed to get the "+
			"instance: %s - Output: %s - Err: %s", err.Error(),
			outb.String(), errb.String()), w, r, store)
		return
	}
	output := outb.String()

	match := darcMatch.FindStringSubmatch(output)
	// at this point we expect match to contain the full match 'DarcID: aef123'
	// and the capturing group 'aef123'
	if len(match) != 2 {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf(
			"number of match '%d' != 2 for the following string: '%s', got "+
				"this match slice: %s", len(match), output, match), w, r, store)
		return
	}
	darcID := match[1]

	purposesStr := ""
	usesStr := ""

	log.Info("getting the darc info")
	cmd = new(exec.Cmd)
	cmd = exec.Command("./bcadmin", "darc", "show", "--darc", darcID, "--bc",
		session.BcPath)

	log.Info(fmt.Sprintf("command created: %s", cmd.Args))
	outb.Reset()
	errb.Reset()
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf("failed to get the "+
			"darc show: %s - Output: %s - Err: %s", err.Error(),
			outb.String(), errb.String()), w, r, store)
		return
	}
	output = outb.String()
	log.Info("here is the output: ", output)

	darcStr := output

	if !conf.Standalone {
		match = readAttrRule.FindStringSubmatch(darcStr)
		// at this point we expect match to contain the full match and the
		// capturing group. Note that this do not handle the case where there is
		// no 'attr:allowed' rule.
		if len(match) != 2 {
			xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf(
				"number of match '%d' != 2 for the following string: '%s', got "+
					"this match slice: %s", len(match), output, match), w, r, store)
			return
		}
		attrExpression := match[1]
		vals, err := url.ParseQuery(attrExpression)
		if err != nil {
			xhelpers.RedirectWithErrorFlash("/", "failed to parse query: "+
				err.Error(), w, r, store)
			return
		}

		purposesStr = vals.Get("purposes")
		usesStr = vals.Get("uses")
	}

	log.Info("purposes: ", purposesStr)
	log.Info("uses: ", usesStr)

	purposes := strings.Split(purposesStr, ",")
	log.Info("here are the Darc purposes: ", purposes)
	uses := strings.Split(usesStr, ",")
	log.Info("here are the Darc uses: ", uses)

	// we have everything we need, let's render the view

	type viewData struct {
		Title    string
		Flash    []xhelpers.Flash
		Session  *models.Session
		Dataset  *catalogc.Dataset
		DarcID   string
		DarcStr  string
		Uses     []string
		Purposes []string
	}

	t, err := template.New("template").Funcs(template.FuncMap{
		"addCheck": func(els []string, target string) string {
			for _, el := range els {
				if el == target {
					return "checked"
				}
			}
			return ""
		},
	}).ParseFiles("views/layout.gohtml", "views/datasets/show.gohtml")
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
		Title:    "Your datasets",
		Flash:    flashes,
		Session:  session,
		Dataset:  dataset,
		DarcID:   darcID,
		DarcStr:  darcStr,
		Uses:     uses,
		Purposes: purposes,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf(
			"Error while executing template: %s\n", err.Error()), w, r, store)
		return
	}
}

func datasetsShowPut(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {
	// we assume that r.ParseForm() has already been called.

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		xhelpers.RedirectWithErrorFlash("/", "failed to get the dataset id in url", w, r, store)
		return
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

	missing := ""
	title := r.PostFormValue("title")
	if title == "" {
		missing += "- 'title' is empty "
	}
	description := r.PostFormValue("description")
	if description == "" {
		missing += "- 'description' is empty "
	}
	cloudURL := r.PostFormValue("cloudURL")
	if cloudURL == "" {
		missing += "- 'cloud URL' is empty "
	}
	sha2 := r.PostFormValue("sha2")
	if sha2 == "" {
		missing += "- 'sha2' is empty "
	}
	if missing != "" {
		xhelpers.RedirectWithErrorFlash("/datasets/new",
			"Error, some fields were empty: "+missing, w, r, store)
	}

	// creating the task
	tef := xhelpers.NewTaskEventFactory("DO Manager")
	task := xhelpers.NewTask(fmt.Sprintf("Update infos for dataset '%s'", title))
	task.GoBackLink = `<p><a class="pure-button" href="/datasets">🔙 Back to the list of datasets</a></p>`
	task.AddInfof(tef.Source, "starting the update process", "got this POST form: %v", r.PostForm)

	go func() {

		identityStr := session.Cfg.AdminIdentity.String()
		cmd := exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
			"catalog", "invoke", "updateDataset", "--bc", session.BcPath,
			"--instid", conf.CatalogID, "--title", title, "--description",
			description, "--cloudURL", cloudURL, "--sha2", sha2, "--identityStr",
			identityStr, "--calypsoWriteID", id)
		task.AddInfof(tef.Source, "saving the attributes on the catalog", "using this command: %v", cmd.Args)
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "command failed", fmt.Sprintf("failed to update "+
				"the dataset: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}
		output := outb.String()

		task.CloseOK(tef.Source, "dataset's infos updated", output)

	}()

	xhelpers.RedirectWithInfoFlash(fmt.Sprintf("/showtasks/%d", task.Index),
		fmt.Sprintf("Task to update dataset's info with index %d created", task.Index), w, r, store)
}

// This function should not be called be the dataset owner, it is here for debug
// purpose
func datasetsShowDelete(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {
	// we assume that r.ParseForm() has already been called.

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		xhelpers.RedirectWithErrorFlash("/", "failed to get the dataset id in url", w, r, store)
		return
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

	datasetTitle := r.PostFormValue("datasetTitle")
	if datasetTitle == "" {
		datasetTitle = id
	}

	// creating the task
	tef := xhelpers.NewTaskEventFactory("DO Manager")
	task := xhelpers.NewTask(fmt.Sprintf("Deletion of the dataset '%s'", datasetTitle))
	task.GoBackLink = `<p><a class="pure-button" href="/datasets">🔙 Back to the list of datasets</a></p>`
	task.AddInfof(tef.Source, "starting the deletion process", "got this POST form: %v", r.PostForm)

	go func() {

		// Get the dataset

		identityStr := session.Cfg.AdminIdentity.String()
		cmd := exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
			"catalog", "getDatasets", "-i", conf.CatalogID, "-bc", session.BcPath,
			"-identityStr", identityStr, "--toJson")
		task.AddInfof(tef.Source, "getting the dataset from the catalog", "using the following command: %v", cmd.Args)
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "command failed", fmt.Sprintf("failed to get the "+
				"list of datasets: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}
		datasets := make([]*catalogc.Dataset, 0)
		err = json.Unmarshal(outb.Bytes(), &datasets)
		if err != nil {
			task.CloseError(tef.Source, "failed to unmarshal result",
				err.Error())
			return
		}

		if len(datasets) == 1 && datasets[0] == nil {
			datasets = []*catalogc.Dataset{}
		}

		var dataset *catalogc.Dataset
		for _, d := range datasets {
			if d.CalypsoWriteID == id {
				dataset = d
				break
			}
		}

		if dataset == nil {
			task.CloseError(tef.Source, "dataset is nil", fmt.Sprintf(
				"dataset with id '%s' not found for identityStr '%s'",
				id, identityStr))
			return
		}

		// Remove the dataset from the cloud (if found)

		task.AddInfof(tef.Source, "removing the dataset from the cloud", "using the following filename: %s", filepath.Base(dataset.CloudURL))
		minioClient, err := xhelpers.GetMinioClient()
		if err != nil {
			task.CloseError(tef.Source, "failed to get minio client",
				err.Error())
		}
		err = minioClient.RemoveObject("datasets", filepath.Base(dataset.CloudURL))
		minioErr := minio.ToErrorResponse(err)
		if err != nil && minioErr.Code == "NoSuchKey" {
			task.AddInfof(tef.Source, "dataset not found on the cloud", "Object 'datasets/%s' not found, nothing to delete on the "+
				"cloud storage then", filepath.Base(dataset.CloudURL))
		} else if err != nil {
			task.CloseError(tef.Source, "failed delete object on the "+
				"cloud", err.Error())
			return
		}

		// Delete dataset from the catalog

		cmd = new(exec.Cmd)
		cmd = exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
			"catalog", "invoke", "deleteDataset", "--bc", session.BcPath,
			"--instid", conf.CatalogID, "--identityStr", identityStr,
			"--calypsoWriteID", id)
		task.AddInfof(tef.Source, "removing the dataset from the catalog", "using the following command: %v", cmd.Args)
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		outb.Reset()
		errb.Reset()
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "failed to launch the command",
				fmt.Sprintf("failed to delete the dataset: %s - Output: %s - "+
					"Err: %s", err.Error(), outb.String(), errb.String()))
			return
		}
		output := outb.String()
		task.CloseOK(tef.Source, "dataset deleted", output)
	}()

	xhelpers.RedirectWithInfoFlash(fmt.Sprintf("/showtasks/%d", task.Index),
		fmt.Sprintf("Task to delete dataset with index %d created", task.Index),
		w, r, store)
}

func datasetsShowAttributesPut(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {
	// we assume that r.ParseForm() has already been called.

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		xhelpers.RedirectWithErrorFlash("/", "failed to get the dataset id "+
			"in url", w, r, store)
		return
	}

	datasetTitle := r.PostFormValue("datasetTitle")
	if datasetTitle == "" {
		datasetTitle = id
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

	log.Lvlf1("We got this post form: %v", r.PostForm)

	// creating the task
	tef := xhelpers.NewTaskEventFactory("DO Manager")
	task := xhelpers.NewTask(fmt.Sprintf("Update attributes on dataset '%s'", datasetTitle))
	task.GoBackLink = `<p><a class="pure-button" href="/datasets">🔙 Back to the list of datasets</a></p>`
	task.AddInfof(tef.Source, "starting the update process", "got this POST form: %v", r.PostForm)

	go func() {

		// Let's get the original metadata, update it with what we got from the post
		// form and then save it back to the catalog.

		identityStr := session.Cfg.AdminIdentity.String()
		cmd := exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
			"catalog", "getDatasets", "-i", conf.CatalogID, "-bc", session.BcPath,
			"-identityStr", identityStr, "--toJson")
		task.AddInfof(tef.Source, "getting the original metadata", "using this command: %v", cmd.Args)
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "command failed", fmt.Sprintf("failed to get the "+
				"list of datasets: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}
		datasets := make([]*catalogc.Dataset, 0)
		err = json.Unmarshal(outb.Bytes(), &datasets)
		if err != nil {
			task.CloseError(tef.Source, "failed to unmarshal result", err.Error())
			return
		}

		if len(datasets) == 1 && datasets[0] == nil {
			datasets = []*catalogc.Dataset{}
		}

		var dataset *catalogc.Dataset
		for _, d := range datasets {
			if d.CalypsoWriteID == id {
				dataset = d
				break
			}
		}

		if dataset == nil {
			task.CloseError(tef.Source, "dataset not found", fmt.Sprintf(
				"dataset with id '%s' not found for identityStr '%s'",
				id, identityStr))
			return
		}

		metadata := dataset.Metadata

		if metadata == nil {
			task.CloseError(tef.Source, "This dataset has a nil "+
				"Metadata field on the catalog", "")
			return
		}

		// We reset the metadata and set the attributes based on what is in the form

		task.AddInfo(tef.Source, "setting the attributes from the form on the metadata",
			"resetting the original dataset and calling tryUpdate on each attribute from the POST form")

		metadata.Reset()
		for key, values := range r.PostForm {
			for _, val := range values {
				metadata.TryUpdate(key, val)
			}
		}

		metadataJSONBuf, err := json.Marshal(metadata)
		if err != nil {
			task.CloseError(tef.Source, "failed to mnarshal metadata "+
				"to JSON", err.Error())
			return
		}
		metadataJSONStr := string(metadataJSONBuf)

		cmd = new(exec.Cmd)
		cmd = exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
			"catalog", "invoke", "updateDataset", "--bc", session.BcPath,
			"--instid", conf.CatalogID, "--metadataJSON", metadataJSONStr,
			"--identityStr", identityStr, "--calypsoWriteID", id)
		task.AddInfof(tef.Source, "saving the new metadata on the catalog", "using this command: %v", cmd.Args)
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		outb.Reset()
		errb.Reset()
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "command failed", fmt.Sprintf("failed to update "+
				"the dataset: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}

		// Now let's find the DARC associated with this write instance. For this we
		// first need the "bcadmin instance get" command, then the "bcadmin darc
		// show"

		log.Info("getting the darc ID")
		cmd = new(exec.Cmd)
		cmd = exec.Command("./bcadmin", "instance", "get", "-i", id, "--bc",
			session.BcPath)
		task.AddInfof(tef.Source, "getting the associated DARC ID", "using this command: %v", cmd.Args)
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		outb.Reset()
		errb.Reset()
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "command failed", fmt.Sprintf("failed to get the "+
				"instance: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}
		output := outb.String()

		match := darcMatch.FindStringSubmatch(output)
		// at this point we expect match to contain the full match 'DarcID: aef123'
		// and the capturing group 'aef123'
		if len(match) != 2 {
			task.CloseError(tef.Source, fmt.Sprintf(
				"number of match '%d' != 2 for the following string: '%s', got "+
					"this match slice: %s", len(match), output, match), "")
			return
		}
		darcID := match[1]

		log.Info("getting the darc info")

		cmd = new(exec.Cmd)
		cmd = exec.Command("./bcadmin", "darc", "show", "--darc", darcID, "--bc",
			session.BcPath)
		task.AddInfof(tef.Source, "getting the darc info", "using this command: %v", cmd.Args)
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		outb.Reset()
		errb.Reset()
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "command failed", fmt.Sprintf("failed to get the "+
				"darc show: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}
		output = outb.String()
		log.Info("here is the output: ", output)

		darcStr := output

		task.AddInfof(tef.Source, "parsing the spawn:calypsoRead rule", "got this darc string:\n%s", darcStr)
		match = readRule.FindStringSubmatch(darcStr)
		// at this point we expect match to contain the full match and the capturing
		// group.
		if len(match) != 2 {
			task.CloseError(tef.Source, fmt.Sprintf(
				"number of match '%d' != 2 for the following string: '%s', got "+
					"this match slice: %s", len(match), output, match), "")
			return
		}
		// Rule is of form "darc:123 |
		// attr:allowed:uses=clientData&purposes=tata,toto | ed25519:..."
		rule := match[1]
		log.Info("here is the calypso read rule: ", rule)

		match = attrRule.FindStringSubmatch(rule)
		// at this point we expect match to contain the full match and the capturing
		// group. Note that this do not handle the case where there is no
		// 'attr:allowed' rule.
		if len(match) != 2 {
			task.CloseError(tef.Source, fmt.Sprintf(
				"number of match '%d' != 2 for the following string: '%s', got "+
					"this match slice: %s", len(match), output, match), "")
			return
		}

		// attrExpression is of form "uses=clientData&purposes=tata,toto"
		attrExpression := match[1]
		log.Info("here is the attribute to be replaced: ", attrExpression)

		task.AddInfo(tef.Source, "building the new rule expression", "calling metadata.Darc(id)")
		// now let's build the new expression
		newExpression := metadata.Darc(id)

		rule = strings.ReplaceAll(rule, attrExpression, newExpression)
		log.Info("here is the new rule: ", rule)

		// Now its time to call bcadmin
		// Waiting 10 seconds in order to prevent bad counter, because we run an
		// a bcadmin command before.

		cmd = new(exec.Cmd)
		cmd = exec.Command("./bcadmin", "-c", conf.ConfigPath, "darc", "rule",
			"--rule", "spawn:calypsoRead", "--darc", darcID, "--bc", session.BcPath,
			"-id", rule, "-replace", "-restricted")
		task.AddInfof(tef.Source, "updating the DARC", "using this command: %v", cmd.Args)
		time.Sleep(time.Second * 10)
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		outb.Reset()
		errb.Reset()
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "command failed", fmt.Sprintf("failed to update "+
				"the DARC darc show: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}
		output = outb.String()

		task.CloseOK(tef.Source, "dataset's attributes updated", "the last command ran successfully")
	}()

	xhelpers.RedirectWithInfoFlash(fmt.Sprintf("/showtasks/%d", task.Index),
		fmt.Sprintf("Task to update dataset's attributes with index %d created", task.Index), w, r, store)
}

func datasetsShowArchivePut(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {
	// we assume that r.ParseForm() has already been called.

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		xhelpers.RedirectWithErrorFlash("/", "failed to get the dataset id "+
			"in url", w, r, store)
		return
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

	datasetTitle := r.PostFormValue("datasetTitle")
	if datasetTitle == "" {
		datasetTitle = id
	}

	log.Lvlf1("We got this post form: %v", r.PostForm)

	// creating the task
	tef := xhelpers.NewTaskEventFactory("DO Manager")
	task := xhelpers.NewTask(fmt.Sprintf("Remove access from the dataset '%s'", datasetTitle))
	task.GoBackLink = `<p><a class="pure-button" href="/datasets">🔙 Back to the list of datasets</a></p>`
	task.AddInfof(tef.Source, "starting the upload process", "got this POST form: %v", r.PostForm)

	go func() {

		// Get the dataset

		identityStr := session.Cfg.AdminIdentity.String()
		cmd := exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
			"catalog", "getDatasets", "-i", conf.CatalogID, "-bc", session.BcPath,
			"-identityStr", identityStr, "--toJson")
		task.AddInfof(tef.Source, "getting the dataset from the catalog", "using this command: %v", cmd.Args)
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "command failed", fmt.Sprintf("failed to get the "+
				"list of datasets: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}
		datasets := make([]*catalogc.Dataset, 0)
		err = json.Unmarshal(outb.Bytes(), &datasets)
		if err != nil {
			task.CloseError(tef.Source, "failed to unmarshal result: ",
				err.Error())
			return
		}

		if len(datasets) == 1 && datasets[0] == nil {
			datasets = []*catalogc.Dataset{}
		}

		var dataset *catalogc.Dataset
		for _, d := range datasets {
			if d.CalypsoWriteID == id {
				dataset = d
				break
			}
		}

		if dataset == nil {
			task.CloseError(tef.Source, fmt.Sprintf(
				"dataset with id '%s' not found for identityStr '%s'",
				id, identityStr), "")
			return
		}

		// Remove the dataset from the cloud

		task.AddInfof(tef.Source, "removing the encrypted dataset from the cloud", "using this filename: %s", filepath.Base(dataset.CloudURL))

		minioClient, err := xhelpers.GetMinioClient()
		if err != nil {
			task.CloseError(tef.Source, "failed to get minio client: ",
				err.Error())
			return
		}
		err = minioClient.RemoveObject("datasets", filepath.Base(dataset.CloudURL))
		if err != nil {
			task.CloseError(tef.Source, "failed delete object on the "+
				"cloud", err.Error())
			return
		}

		// Update the catalog with the "archived" flag and reset the metadata

		cmd = new(exec.Cmd)
		cmd = exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
			"catalog", "invoke", "archiveDataset", "-i", conf.CatalogID, "-bc",
			session.BcPath, "-identityStr", identityStr, "--calypsoWriteID", id)
		task.AddInfof(tef.Source, "setting the 'archived' flag on the catalog", "using this command: %v", cmd.Args)
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		outb.Reset()
		errb.Reset()
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "command failed", fmt.Sprintf("failed to set the "+
				"dataset in archived state: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}

		// Update the darc by removing the rules concerned by the metadata

		// Now let's find the DARC associated with this write instance. For this we
		// first need the "bcadmin instance get" command, then the "bcadmin darc
		// show"

		log.Info("getting the darc ID")
		cmd = new(exec.Cmd)
		cmd = exec.Command("./bcadmin", "instance", "get", "-i", id, "--bc",
			session.BcPath)
		task.AddInfof(tef.Source, "getting the associated DARC ID", "using this command: %v", cmd.Args)
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		outb.Reset()
		errb.Reset()
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "command failed", fmt.Sprintf("failed to get the "+
				"instance: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}
		output := outb.String()
		match := darcMatch.FindStringSubmatch(output)
		// at this point we expect match to contain the full match 'DarcID: aef123'
		// and the capturing group 'aef123'
		if len(match) != 2 {
			task.CloseError(tef.Source, "command failed", fmt.Sprintf(
				"number of match '%d' != 2 for the following string: '%s', got "+
					"this match slice: %s", len(match), output, match))
			return
		}
		darcID := match[1]

		log.Info("getting the darc info")

		cmd = new(exec.Cmd)
		cmd = exec.Command("./bcadmin", "darc", "show", "--darc", darcID, "--bc",
			session.BcPath)
		task.AddInfof(tef.Source, "getting the DARC info", "using this command: %v", cmd.Args)
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		outb.Reset()
		errb.Reset()
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "commmand failed", fmt.Sprintf("failed to get the "+
				"darc show: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}
		output = outb.String()
		log.Info("here is the output: ", output)

		darcStr := output
		task.AddInfof(tef.Source, "parsing the spawn:calpysoRead rule", "got this DARC sting: %v", darcStr)
		match = readRule.FindStringSubmatch(darcStr)
		// at this point we expect match to contain the full match and the capturing
		// group.
		if len(match) != 2 {
			task.CloseError(tef.Source, fmt.Sprintf(
				"number of match '%d' != 2 for the following string: '%s', got "+
					"this match slice: %s", len(match), output, match), "")
			return
		}
		// Rule is of form "darc:123 |
		// attr:allowed:uses=clientData&purposes=tata,toto | ed25519:..."
		rule := match[1]
		log.Info("here is the calypso read rule: ", rule)

		match = attrRule.FindStringSubmatch(rule)
		// at this point we expect match to contain the full match and the capturing
		// group. Note that this do not handle the case where there is no
		// 'attr:allowed' rule.
		if len(match) != 2 {
			task.CloseError(tef.Source, fmt.Sprintf(
				"number of match '%d' != 2 for the following string: '%s', got "+
					"this match slice: %s", len(match), output, match), "")
			return
		}

		// attrExpression is of form "uses=clientData&purposes=tata,toto"
		attrExpression := match[1]
		log.Info("here is the attribute to be replaced: ", attrExpression)

		rule = strings.ReplaceAll(rule, attrExpression, "( attr:allowed: & attr:must_have: )")
		log.Info("here is the new rule: ", rule)

		// Now its time to call bcadmin
		// Waiting 10 seconds in order to prevent bad counter, because we run an
		// a bcadmin command before.

		cmd = new(exec.Cmd)
		cmd = exec.Command("./bcadmin", "-c", conf.ConfigPath, "darc", "rule",
			"--rule", "spawn:calypsoRead", "--darc", darcID, "--bc", session.BcPath,
			"-id", rule, "-replace", "-restricted")
		task.AddInfof(tef.Source, "setting the new rule", "replacing with this "+
			"expression: '%s', using this command: %v", attrExpression, cmd.Args)
		time.Sleep(time.Second * 10)
		log.Info(fmt.Sprintf("command created: %s", cmd.Args))
		outb.Reset()
		errb.Reset()
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			task.CloseError(tef.Source, "command failed", fmt.Sprintf("failed to update "+
				"the DARC: %s - Output: %s - Err: %s", err.Error(),
				outb.String(), errb.String()))
			return
		}

		task.CloseOK(tef.Source, "access to the dataset removed", "the last command ran successfully")

	}()

	xhelpers.RedirectWithInfoFlash(fmt.Sprintf("/showtasks/%d", task.Index),
		fmt.Sprintf("Task to set dataset in archived mode with index %d created", task.Index), w, r, store)
}

func datasetsShowAuditGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		xhelpers.RedirectWithErrorFlash("/", "failed to get the dataset id "+
			"in url", w, r, store)
		return
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

	cmd := exec.Command("./catadmin", "-c", conf.ConfigPath, "audit",
		"dataset", "-instid", id, "-bc", session.BcPath)

	log.Info(fmt.Sprintf("command created: %s", cmd.Args))
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf("failed to get the "+
			"audit log: %s - Output: %s - Err: %s", err.Error(),
			outb.String(), errb.String()), w, r, store)
		return
	}

	type viewData struct {
		Title     string
		Flash     []xhelpers.Flash
		Session   *models.Session
		AuditHTML string
		ID        string
		ShortID   string
	}

	t, err := template.ParseFiles("views/layout.gohtml", "views/datasets/audit.gohtml")
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
		Title:     "Your datasets",
		Flash:     flashes,
		Session:   session,
		AuditHTML: outb.String(),
		ID:        id,
		ShortID:   id[:8] + "...",
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf(
			"Error while executing template: %s\n", err.Error()), w, r, store)
		return
	}

}

func datasetsShowDebugGet(w http.ResponseWriter, r *http.Request,
	store *sessions.CookieStore, conf *models.Config) {

	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		xhelpers.RedirectWithErrorFlash("/", "failed to get the dataset id "+
			"in url", w, r, store)
		return
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

	identityStr := session.Cfg.AdminIdentity.String()
	cmd := exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
		"catalog", "getDatasets", "-i", conf.CatalogID, "-bc", session.BcPath,
		"-identityStr", identityStr, "--toJson")

	log.Info(fmt.Sprintf("command created: %s", cmd.Args))
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf("failed to get the "+
			"list of datasets: %s - Output: %s - Err: %s", err.Error(),
			outb.String(), errb.String()), w, r, store)
		return
	}
	datasets := make([]*catalogc.Dataset, 0)
	err = json.Unmarshal(outb.Bytes(), &datasets)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", "failed to unmarshal result: "+
			err.Error(), w, r, store)
		return
	}

	if len(datasets) == 1 && datasets[0] == nil {
		datasets = []*catalogc.Dataset{}
	}

	var dataset *catalogc.Dataset
	for _, d := range datasets {
		if d.CalypsoWriteID == id {
			dataset = d
			break
		}
	}

	if dataset == nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf(
			"dataset with id '%s' not found for identityStr '%s'",
			id, identityStr), w, r, store)
		return
	}

	type viewData struct {
		Title   string
		Flash   []xhelpers.Flash
		Session *models.Session
		Dataset *catalogc.Dataset
	}

	t, err := template.New("template").Funcs(template.FuncMap{
		"addCheck": func(els []string, target string) string {
			for _, el := range els {
				if el == target {
					return "checked"
				}
			}
			return ""
		},
	}).ParseFiles("views/layout.gohtml", "views/datasets/debug.gohtml")
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
		Title:   "Your datasets",
		Flash:   flashes,
		Session: session,
		Dataset: dataset,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		xhelpers.RedirectWithErrorFlash("/", fmt.Sprintf(
			"Error while executing template: %s\n", err.Error()), w, r, store)
		return
	}

}
