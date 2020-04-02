package controllers

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"text/template"

	"github.com/dedis/odyssey/catalogc"
	"github.com/dedis/odyssey/dsmanager/app/helpers"
	"github.com/dedis/odyssey/dsmanager/app/models"
	"github.com/gorilla/sessions"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/protobuf"
	bolt "go.etcd.io/bbolt"
)

// DatasetsIndexHandler ...
func DatasetsIndexHandler(gs sessions.Store, conf *models.Config, db *bolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			datasetsGet(w, r, gs, conf)
		}
	}
}

func datasetsGet(w http.ResponseWriter, r *http.Request, store sessions.Store, conf *models.Config) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "false")

	t, err := template.ParseFiles("views/layout.gohtml", "views/datasets.gohtml")
	if err != nil {
		helpers.RedirectWithErrorFlash("/", "Error with template: "+err.Error(), w, r, store)
		return
	}

	type viewData struct {
		Title    string
		Datasets []*catalogc.Dataset
		Flash    []helpers.Flash
	}

	datasets, err := getDatasets(conf)
	if err != nil {
		helpers.AddFlash(w, r, fmt.Sprintf("<pre>Failed to get datasets:\n%s</pre>", err.Error()), store, helpers.Error)
	}

	flashes, err := helpers.ExtractFlash(w, r, store)
	if err != nil {
		log.Errorf("Failed to get flash: %s\n", err.Error())
	}

	p := &viewData{
		Title:    "List of datasets",
		Flash:    flashes,
		Datasets: datasets,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		log.Errorf("Error while executing template: %s\n", err.Error())
	}
}

func getDatasets(conf *models.Config) ([]*catalogc.Dataset, error) {
	cmd := exec.Command("./catadmin", "-c", conf.ConfigPath, "contract",
		"catalog", "get", "-i", conf.CatalogID, "-bc", conf.BCPath,
		"--export")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to get the catalog with id '%s': %s - "+
			"Output: %s - Err: %s", conf.CatalogID, err.Error(), outb.String(),
			errb.String())
	}

	// cmdOut := outb.String()
	// log.LLvl1("here is the output of the value contract:", cmdOut)

	catalog := catalogc.CatalogData{}
	err = protobuf.Decode(outb.Bytes(), &catalog)

	datasets := []*catalogc.Dataset{}

	for _, owner := range catalog.Owners {
		log.LLvl1("getting datasets for owner", owner)
		if owner.Datasets == nil {
			continue
		}
		for _, dataset := range owner.Datasets {
			if dataset.IsArchived {
				continue
			}
			datasets = append(datasets, dataset)
		}
	}

	return datasets, nil
}
