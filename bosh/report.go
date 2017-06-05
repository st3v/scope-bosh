package bosh

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/st3v/scope-bosh/bosh/monit"
)

type pluginSpec struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  string   `json:"api_version"`
}

type report struct {
	ID       string       `json:"ID"`
	Plugins  []pluginSpec `json:"Plugins"`
	Host     hostSpec     `json:"Host"`
	hostname string
}

type hostSpec struct {
	Label             string                          `json:"label"`
	LabelPlural       string                          `json:"label_plural"`
	Nodes             map[string]nodeSpec             `json:"nodes"`
	Shape             string                          `json:"shape"`
	MetadataTemplates map[string]metadataTemplateSpec `json:"metadata_templates"`
	TableTemplates    map[string]tableTemplateSpec    `json:"table_templates"`
}

type metadataTemplateSpec struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Priority int    `json:"priority"`
	From     string `json:"from"`
	Truncate int    `json:"truncate,omitempty"`
}

type tableTemplateSpec struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Prefix string `json:"prefix"`
}

type nodeSpec struct {
	ID       string                `json:"id"`
	Topology string                `json:"topology,omitempty"`
	Latest   map[string]latestSpec `json:"latest,omitempty"`
}

type latestSpec struct {
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`
}

var pluginInfo = pluginSpec{
	ID:          "bosh",
	Label:       "bosh",
	Description: "Reports on Bosh agent properties",
	Interfaces:  []string{"reporter"},
	APIVersion:  "1",
}

func newReport(hostname string) report {
	return report{
		ID:       fmt.Sprintf("%d", rand.Int63()),
		Plugins:  []pluginSpec{pluginInfo},
		Host:     host(),
		hostname: hostname,
	}
}

func (r report) setInstanceSpecData(jobSpec spec) {
	hostID := r.hostID()

	node := nodeSpec{
		ID:       hostID,
		Topology: "host",
		Latest: map[string]latestSpec{
			BoshJobID:         latest(jobSpec.ID),
			BoshJobName:       latest(jobSpec.Job.Name),
			BoshJobIndex:      latest(strconv.Itoa(jobSpec.Index)),
			BoshJobDeployment: latest(jobSpec.Deployment),
		},
	}

	for _, t := range jobSpec.Job.Templates {
		node.Latest[BoshTemplatesPrefix+t.Name] = latest(t.Version)
	}

	for _, p := range jobSpec.Packages {
		node.Latest[BoshPackagesPrefix+p.Name] = latest(p.Version)
	}

	for name, n := range jobSpec.Networks {
		id := BoshNetworksPrefix + name
		label := "Bosh Networks - " + name
		r.Host.TableTemplates[id] = tableTemplateSpec{ID: id, Label: label, Prefix: id}

		node.Latest[id+"ip"] = latest(n.IP)
		node.Latest[id+"gateway"] = latest(n.Gateway)
		node.Latest[id+"netmask"] = latest(n.Netmask)
		node.Latest[id+"default"] = latest(strings.Join(n.Default, ", "))
	}

	r.Host.Nodes[hostID] = node
}

func (r report) hostID() string {
	return fmt.Sprintf("%s;<host>", r.hostname)
}

func (r report) setMonitProcesses(processes []monit.Process) {
	hostID := r.hostID()
	node := r.Host.Nodes[hostID]

	for k := range node.Latest {
		if strings.HasPrefix(k, MonitProcessesPrefix) {
			delete(node.Latest, k)
		}
	}

	for _, p := range processes {
		node.Latest[MonitProcessesPrefix+p.Name()] = latest(p.Status())
	}

	r.Host.Nodes[hostID] = node
}

func host() hostSpec {
	return hostSpec{
		Label:          "host",
		LabelPlural:    "hosts",
		Shape:          "circle",
		Nodes:          map[string]nodeSpec{},
		TableTemplates: tableTemplates,
	}
}

func latest(v string) latestSpec {
	return latestSpec{
		Timestamp: time.Now(),
		Value:     v,
	}
}

const (
	BoshJobPrefix        = "bosh_job_"
	BoshTemplatesPrefix  = "bosh_templates_"
	BoshPackagesPrefix   = "bosh_packages_"
	BoshNetworksPrefix   = "bosh_networks_"
	BoshJobName          = "bosh_job_name"
	BoshJobID            = "bosh_job_id"
	BoshJobIndex         = "bosh_job_index"
	BoshJobDeployment    = "bosh_job_deployment"
	MonitProcessesPrefix = "monit_processes_"
)

var tableTemplates = map[string]tableTemplateSpec{
	BoshJobPrefix:        {ID: BoshJobPrefix, Label: "Bosh Job Info", Prefix: BoshJobPrefix},
	BoshTemplatesPrefix:  {ID: BoshTemplatesPrefix, Label: "Bosh Templates", Prefix: BoshTemplatesPrefix},
	BoshPackagesPrefix:   {ID: BoshPackagesPrefix, Label: "Bosh Packages", Prefix: BoshPackagesPrefix},
	MonitProcessesPrefix: {ID: MonitProcessesPrefix, Label: "Monit Processes", Prefix: MonitProcessesPrefix},
}
