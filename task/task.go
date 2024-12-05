package task

import "github.com/google/uuid"

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
)

func (s State) String() string {
	return [...]string{
		"Pending",
		"Scheduled",
		"Running",
		"Completed",
		"Failed",
	}[s]
}

type Task struct {
	ID    uuid.UUID
	Name  string
	State State
}
