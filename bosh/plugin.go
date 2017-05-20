package bosh

import (
	"encoding/json"
	"log"
	"net/http"
)

type plugin struct {
	report report
}

func NewPlugin(hostname, jobSpecPath string) *plugin {
	report := newReport(hostname)

	if jobSpec, err := loadSpec(jobSpecPath); err != nil {
		log.Printf("error loading job spec from %q: %v\n", jobSpecPath, err)
	} else {
		report.AddJob(jobSpec)
	}

	return &plugin{report}
}

func (p *plugin) Report(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(p.report); err != nil {
		log.Printf("error encoding report: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
