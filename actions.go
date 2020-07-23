package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/qri-io/apiutil"
	"github.com/qri-io/dataset"
	"github.com/qri-io/qri/base"
	"github.com/qri-io/qri/lib"
	reporef "github.com/qri-io/qri/repo/ref"
)

// SimActionHandler triggers actions on this server that simulate real-world
// behaviour
func SimActionHandler(inst *lib.Instance) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.ToLower(r.FormValue("action"))
		act, ok := simActions[key]
		if !ok {
			apiutil.WriteErrResponse(w, http.StatusBadRequest, fmt.Errorf("action not found: '%s'", key))
			return
		}

		if err := act(r.Context(), inst); err != nil {
			log.Errorf("running action %s: %s", key, err)
			apiutil.WriteErrResponse(w, http.StatusInternalServerError, err)
			return
		}

		apiutil.WriteResponse(w, "ok")
	}
}

// simActions is a mapping of simulation actions keyed by string
var simActions = map[string]simActionFunc{
	"createsynthsdataset": createSynthsDataset,
	"appendsynthsdataset": appendSynthsDataset,
}

type simActionFunc func(ctx context.Context, inst *lib.Instance) error

func createSynthsDataset(ctx context.Context, inst *lib.Instance) error {
	dsm := lib.NewDatasetMethods(inst)
	res := reporef.DatasetRef{}
	err := dsm.Save(&lib.SaveParams{
		Ref: "me/synths",
		Dataset: &dataset.Dataset{
			Meta: &dataset.Meta{
				Title:       "synthesizers",
				Description: "A list of great types of synthesizers",
			},
			BodyPath: "body.csv",
			BodyBytes: []byte(`company,name,year_of_release,initial_cost,initial_cost_adjusted
moog,little phatty,,,,
moog,sub 37,,,,
moog,subsequent 37,,,,
`),
		},
	}, &res)

	if err != nil {
		log.Errorf("createSynthsDataset: error saving dataset: %s", res)
		return err
	}

	res.Published = true
	err = base.SetPublishStatus(inst.Repo(), &res, true)
	if err != nil {
		log.Errorf("createSynthsDataset: error setting published status: %s", res)
		return err
	}
	log.Infof("createSynthsDataset dataset saved: %s", res)
	return nil
}

func appendSynthsDataset(ctx context.Context, inst *lib.Instance) error {
	dsm := lib.NewDatasetMethods(inst)
	res := reporef.DatasetRef{}
	err := dsm.Save(&lib.SaveParams{
		Ref: "me/synths",
		Dataset: &dataset.Dataset{
			BodyPath: "body.csv",
			BodyBytes: []byte(`company,name,year_of_release,initial_cost,initial_cost_adjusted
moog,little phatty,,,,
moog,sub 37,,,,
moog,subsequent 37,,,,
novation,bass station,,,,
`),
		},
	}, &res)

	if err != nil {
		log.Errorf("appendSynthsDataset: error saving dataset: %s", res)
		return err
	}

	res.Published = true
	err = base.SetPublishStatus(inst.Repo(), &res, true)
	if err != nil {
		log.Errorf("appendSynthsDataset: error setting published status: %s", res)
		return err
	}
	log.Infof("appendSynthsDataset saved: %s", res)
	return nil
}
