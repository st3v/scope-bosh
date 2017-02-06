package bosh

import (
	"encoding/json"
	"os"
)

type spec struct {
	Job        job                `json:"job"`
	Packages   map[string]pkg     `json:"packages"`
	Networks   map[string]network `json:"networks"`
	Deployment string             `json:"deployment"`
	ID         string             `json:"id"`
	Index      int                `json:"index"`
}

type job struct {
	Name      string     `json:"name"`
	Templates []template `json:"templates"`
}

type template struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type pkg struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type network struct {
	CloudProperties map[string]string `json:"cloud_properties"`
	Default         []string          `json:"default"`
	Gateway         string            `json:"gateway"`
	IP              string            `json:"ip"`
	Netmask         string            `json:"netmask"`
}

func loadSpec(path string) (spec, error) {
	result := spec{}

	file, err := os.Open(path)
	if err != nil {
		return result, err
	}

	if err := json.NewDecoder(file).Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}
