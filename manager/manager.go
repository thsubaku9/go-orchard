package manager

import (
	"fmt"
	"orchard/task"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Manager struct {
	Pending       queue.Queue
	TaskDb        map[string][]uuid.UUID
	EventDbmap    map[uuid.UUID]string
	Workers       []string
	WorkerTaskMap map[string][]*task.Task
	TaskWorkerMap map[string][]*task.Task
}

func (m *Manager) SelectWorker() {
	fmt.Println("I will select an appropriate worker")
}
func (m *Manager) UpdateTasks() {
	fmt.Println("I will update tasks")
}
func (m *Manager) SendWork() {
	fmt.Println("I will send work to workers")
}
