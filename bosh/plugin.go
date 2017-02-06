package bosh

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type plugin struct {
	lock     sync.RWMutex
	hostname string
	specPath string
}

func NewPlugin(hostname, jobSpecPath string) *plugin {
	return &plugin{
		hostname: hostname,
		specPath: jobSpecPath,
	}
}

func (p *plugin) Report(w http.ResponseWriter, r *http.Request) {
	p.lock.Lock()
	defer p.lock.Unlock()

	report := newReport(p.hostname)

	if jobSpec, err := loadSpec(p.specPath); err != nil {
		log.Printf("error loading job spec from %q: %v\n", p.specPath, err)
	} else {
		report.AddJob(jobSpec)
	}

	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("error encoding report: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
