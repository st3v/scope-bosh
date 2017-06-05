package monit

type Client interface {
	GetProcesses() ([]Process, error)
}
