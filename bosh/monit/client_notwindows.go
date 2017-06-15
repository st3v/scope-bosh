// +build !windows

package monit

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	boshmonit "github.com/cloudfoundry/bosh-agent/jobsupervisor/monit"
	boshhttp "github.com/cloudfoundry/bosh-utils/http"
	boshlogger "github.com/cloudfoundry/bosh-utils/logger"
)

const (
	shortRetryStrategyAttempts = uint(20)
	longRetryStrategyAttempts  = uint(300)
	retryDelay                 = 1 * time.Second
	host                       = "127.0.0.1:2822"
	group                      = "vcap"
	credentialsFile            = "/var/vcap/monit/monit.user"
)

type client struct {
	boshMonitClient boshmonit.Client
}

func NewClient() (*client, error) {
	logger := boshlogger.NewLogger(boshlogger.LevelNone)
	httpClient := http.DefaultClient

	shortHTTPClient := boshhttp.NewRetryClient(
		httpClient,
		shortRetryStrategyAttempts,
		retryDelay,
		logger,
	)

	longHTTPClient := boshmonit.NewMonitRetryClient(
		httpClient,
		longRetryStrategyAttempts,
		shortRetryStrategyAttempts,
		retryDelay,
		logger,
	)

	user, password, err := getCredentials()
	if err != nil {
		return nil, err
	}

	boshMonitClient := boshmonit.NewHTTPClient(
		host,
		user,
		password,
		shortHTTPClient,
		longHTTPClient,
		logger,
	)

	return &client{boshMonitClient}, nil
}

func (c *client) GetProcesses() ([]Process, error) {
	status, err := c.boshMonitClient.Status()
	if err != nil {
		return nil, err
	}

	services := status.ServicesInGroup(group)
	processes := make([]Process, len(services))
	for i, s := range services {
		processes[i] = process{
			name:   s.Name,
			status: s.Status,
		}
	}
	return processes, nil
}

func getCredentials() (string, string, error) {
	contents, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		return "", "", err
	}

	credParts := strings.SplitN(string(contents), ":", 2)
	if len(credParts) != 2 {
		return "", "", errors.New("Malformatted monit user file, expecting username and passsword separated by ':'")
	}

	return credParts[0], credParts[1], nil
}
