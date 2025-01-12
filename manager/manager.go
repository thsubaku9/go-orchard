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

func (m *Manager) AddTask(te task.TaskEvent) {
	m.Pending.Enqueue(te)
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

	for _, worker := range m.Workers {
		log.Printf("Checking worker %v for task updates", worker)
		url := fmt.Sprintf("http://%s/tasks", worker)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error connecting to %v: %v\n", worker, err)
			return
		} else if resp.StatusCode != http.StatusOK {
			log.Printf("Error sending request: %v\n", err)
			return
		}
		d := json.NewDecoder(resp.Body)
		var tasks []task.Task
		err = d.Decode(&tasks)
		if err != nil {
			log.Printf("Error unmarshalling tasks: %s\n", err.Error())
		}
		for _, t := range tasks {
			log.Printf("Attempting to update task %v\n", t.ID)
			_, ok := m.TaskDb[t.ID]
			if !ok {
				log.Printf("Task with ID %s not found\n", t.ID)
				return
			}
			if m.TaskDb[t.ID].State != t.State {
				m.TaskDb[t.ID].State = t.State
			}
			m.TaskDb[t.ID].StartTime = t.StartTime
			m.TaskDb[t.ID].FinishTime = t.FinishTime
			m.TaskDb[t.ID].ContainerId = t.ContainerId
		}
	}

}

func New(workers []string) *Manager {

	taskDb := make(map[uuid.UUID]*task.Task)
	eventDb := make(map[uuid.UUID]*task.TaskEvent)
	workerTaskMap := make(map[string][]uuid.UUID)
	taskWorkerMap := make(map[uuid.UUID]string)
	for worker := range workers {
		workerTaskMap[workers[worker]] = []uuid.UUID{}
	}

	return &Manager{
		Pending:       *queue.New(),
		TaskDb:        taskDb,
		EventDb:       eventDb,
		Workers:       workers,
		WorkerTaskMap: workerTaskMap,
		TaskWorkerMap: taskWorkerMap,
	}
}
