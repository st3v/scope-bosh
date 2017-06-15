package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/st3v/scope-bosh/bosh"
	"github.com/st3v/scope-bosh/bosh/monit"
)

var (
	pluginsRoot     string
	hostname        string
	jobSpecPath     string
	refreshInterval time.Duration
)

func init() {
	flag.StringVar(
		&pluginsRoot,
		"pluginsRoot",
		"/var/run/scope/plugins",
		"root directory for scope plugin sockets",
	)

	flag.StringVar(
		&jobSpecPath,
		"jobSpec",
		"/var/vcap/bosh/spec.json",
		"path to Bosh job spec",
	)

	flag.StringVar(
		&hostname,
		"hostname",
		"",
		"hostname as reported by scope",
	)

	flag.DurationVar(
		&refreshInterval,
		"refreshInterval",
		3*time.Second,
		"interval to fetch for process status updates from monit",
	)
}

func main() {
	flag.Parse()

	if hostname == "" {
		var err error
		hostname, err = os.Hostname()
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Starting on %s...\n", hostname)

	socket := filepath.Join(pluginsRoot, "bosh", "bosh.sock")

	listener, err := listen(socket)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		listener.Close()
		os.RemoveAll(filepath.Dir(socket))
	}()

	handleSignals()

	monitClient, err := monit.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	plugin := bosh.NewPlugin(hostname, jobSpecPath, monitClient, refreshInterval)
	defer plugin.Close()

	http.HandleFunc("/report", plugin.Report)

	log.Fatal(http.Serve(listener, nil))
}

func listen(socket string) (net.Listener, error) {
	os.RemoveAll(filepath.Dir(socket))
	if err := os.MkdirAll(filepath.Dir(socket), 0700); err != nil {
		return nil, fmt.Errorf(
			"error creating directory %q: %v",
			filepath.Dir(socket),
			err,
		)
	}

	listener, err := net.Listen("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("error listening on %q: %v", socket, err)
	}

	log.Printf("Listening on unix://%s\n", socket)
	return listener, nil
}

func handleSignals() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interrupt
		os.Exit(0)
	}()
}
