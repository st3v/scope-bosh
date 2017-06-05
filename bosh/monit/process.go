package monit

type Process interface {
	Name() string
	Status() string
}

type process struct {
	name   string
	status string
}

func (p process) Name() string {
	return p.name
}

func (p process) Status() string {
	return p.status
}
