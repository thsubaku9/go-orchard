package task

import (
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type State int

type Event int

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
	Dropped
)

const (
	SpinUp Event = iota
	SpinDown
)

func (s State) String() string {
	return [...]string{
		"Pending",
		"Scheduled",
		"Running",
		"Completed",
		"Failed",
		"Dropped",
	}[s]
}

var TaskFSM = FSM[State, Event]{
	transitionListing: map[State][]State{
		Pending:   {Scheduled},
		Scheduled: {Scheduled, Running, Failed},
		Running:   {Running, Completed, Failed},
		Completed: {},
		Dropped:   {},
	},
	nextMapping: map[State]map[Event]State{
		Pending: {
			SpinUp: Scheduled,
		},
		Scheduled: {
			SpinUp:   Running,
			SpinDown: Completed,
		},
		Running: {
			SpinDown: Completed,
		},
	},

	mappingMissingState: Dropped,
}

type Task struct {
	ID            uuid.UUID
	Name          string
	State         State
	Event         Event
	Image         string
	CPU           float64
	Memory        int
	Disk          int
	ExposedPorts  nat.PortSet
	HostPorts     nat.PortMap
	PortBindings  map[string]string
	StartTime     time.Time
	FinishTime    time.Time
	ContainerId   string
	TaskConfig    Config
	RestartPolicy string
	HealthCheck   string
	RestartCount  int
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
