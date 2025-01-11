package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"orchard/task"
	"orchard/worker"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Manager struct {
	Pending       queue.Queue
	TaskDb        map[uuid.UUID]*task.Task
	EventDb       map[uuid.UUID]*task.TaskEvent
	Workers       []string
	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
	LastWorker    int
}

func (m *Manager) SelectWorker() string {
	workerId := m.Workers[m.LastWorker]

	m.LastWorker += 1
	m.LastWorker = m.LastWorker % len(m.Workers)
	return workerId
}

func (m *Manager) SendWork() {
	if m.Pending.Len() == 0 {
		return
	}

	workerId := m.SelectWorker()
	te := m.Pending.Dequeue().(task.TaskEvent)
	log.Printf("Pulled %v off pending queue\n", te.Task)
	m.EventDb[te.ID] = &te

	m.WorkerTaskMap[workerId] = append(m.WorkerTaskMap[workerId], te.Task.ID)
	m.TaskWorkerMap[te.Task.ID] = workerId
	te.Task.State = task.Scheduled

	m.TaskDb[te.Task.ID] = &te.Task

	data, err := json.Marshal(te)
	if err != nil {
		log.Printf("Unable to marshal taskevent object: %v.\n", te)
	}

	url := fmt.Sprintf("http://%s/tasks", workerId)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error connecting to %v: %v\n", workerId, err)
		m.Pending.Enqueue(te)
		return
	}

	e := worker.StandardResponse{}
	json.NewDecoder(resp.Body).Decode(&e)

	if resp.StatusCode != http.StatusCreated {
		log.Printf("Error: %s", e.ErrorMsg)
		return
	}

	var taskResult task.Task = e.Response.(task.Task)

	log.Printf("%#v\n", taskResult)
}

func (m *Manager) UpdateTasks() {
	fmt.Println("I will update tasks")
}
