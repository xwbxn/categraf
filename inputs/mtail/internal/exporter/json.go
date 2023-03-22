// Copyright 2015 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package exporter

import (
	"encoding/json"
	"expvar"
	"log"
	"net/http"
)

var exportJSONErrors = expvar.NewInt("exporter_json_errors")

// HandleJSON exports the metrics in JSON format via HTTP.
func (e *Exporter) HandleJSON(w http.ResponseWriter, r *http.Request) {
	b, err := json.MarshalIndent(e.store, "", "  ")
	if err != nil {
		exportJSONErrors.Add(1)
		log.Printf("error marshalling metrics into json:%s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	if _, err := w.Write(b); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
