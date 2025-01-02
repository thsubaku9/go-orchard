package task

import (
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type State int

type Event string

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

var TaskFSM = FSM[State, Event]{
	transitionListing: map[State][]State{
		Pending:   {Scheduled},
		Scheduled: {Scheduled, Running, Failed},
		Running:   {Running, Completed, Failed},
		Completed: {},
		Failed:    {},
	},
}

type Task struct {
	ID            uuid.UUID
	Name          string
	State         State
	Image         string
	Memory        int
	Disk          int
	ExposedPorts  nat.PortSet
	PortBindings  map[string]string
	RestartPolicy string
	StartTime     time.Time
	FinishTime    time.Time
	ContainerId   string
	TaskConfig    Config
}

func NewConfig(t *Task) Config {
	if t == nil {
		return Config{}
	}

	return t.TaskConfig
}

type TaskEvent struct {
	ID        uuid.UUID
	State     State
	Timestamp time.Time
	Task      Task
}
