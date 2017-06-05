package bosh

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/st3v/scope-bosh/bosh/monit"
)

type plugin struct {
	report      report
	monitClient monit.Client
	jobSpecPath string
	lock        sync.RWMutex
	hostname    string
	done        chan struct{}
	err         error
}

func NewPlugin(hostname, jobSpecPath string, monitClient monit.Client, refreshInterval time.Duration) *plugin {
	p := &plugin{
		report:      newReport(hostname),
		monitClient: monitClient,
		jobSpecPath: jobSpecPath,
		hostname:    hostname,
		done:        make(chan struct{}),
	}

	p.refreshReport(refreshInterval)

	return p
}

func (p *plugin) Report(w http.ResponseWriter, r *http.Request) {
	if err := p.getErr(); err != nil {
		log.Printf("error fetching report: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(p.getReport()); err != nil {
		log.Printf("error encoding report: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (p *plugin) Close() {
	close(p.done)
}

func (p *plugin) getReport() report {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.report
}

func (p *plugin) setReport(r report) {
	p.lock.Lock()
	p.report = r
	p.err = nil
	p.lock.Unlock()
}

func (p *plugin) getErr() error {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.err
}

func (p *plugin) setErr(err error) {
	p.lock.Lock()
	p.err = err
	p.lock.Unlock()
}

func (p *plugin) refreshReport(refreshInterval time.Duration) {
	go func() {
		for {
			select {
			case <-time.After(refreshInterval):
				r := newReport(p.hostname)

				if jobSpec, err := loadSpec(p.jobSpecPath); err != nil {
					log.Printf("error loading job spec from %q: %v\n", p.jobSpecPath, err)
					p.setErr(err)
					continue
				} else {
					r.setInstanceSpecData(jobSpec)
				}

				if processes, err := p.monitClient.GetProcesses(); err != nil {
					log.Printf("error getting processes from monit: %v\n", err)
					p.setErr(err)
					continue
				} else {
					r.setMonitProcesses(processes)
				}

				p.setReport(r)
			case <-p.done:
				return
			}
		}
	}()
}
