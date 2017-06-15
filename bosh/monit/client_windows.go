// +build windows

package monit

type client struct{}

func NewClient() (client, error) {
	return client{}, nil
}

func (c client) GetProcesses() ([]Process, error) {
	return []Process{}, nil
}
