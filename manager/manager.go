package manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"orchard/api"
	"orchard/node"
	"orchard/scheduler"
	"orchard/task"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Manager struct {
	Pending       queue.Queue
	TaskDb        map[uuid.UUID]*task.Task
	EventDb       map[uuid.UUID]*task.TaskEvent
	Workers       []string
	WorkerIpMap   map[string]string
	WorkerTaskMap map[string]map[uuid.UUID]interface{}
	TaskWorkerMap map[uuid.UUID]string
	LastWorker    int
	WorkerNodes   []*node.Node
	Scheduler     scheduler.Scheduler
}

func (m *Manager) AddTask(te task.TaskEvent) {
	m.Pending.Enqueue(te)
}

func (m *Manager) SelectWorker(t task.Task) (*node.Node, error) {
	candidateNodes := m.Scheduler.SelectCandidateNodes(t, m.WorkerNodes)
	if candidateNodes == nil {
		msg := fmt.Sprintf("No available candidates match resource request for task %v", t.ID)
		err := errors.New(msg)
		return nil, err
	}
	scores := m.Scheduler.ScoreNodes(t, candidateNodes)
	return m.Scheduler.PickNode(scores, candidateNodes), nil
}

func (m *Manager) SendWork() {
	if m.Pending.Len() == 0 {
		return
	}

	te := m.Pending.Dequeue().(task.TaskEvent)
	log.Printf("Pulled %v off pending queue\n", te.Task)
	m.EventDb[te.ID] = &te

	taskWorker, ok := m.TaskWorkerMap[te.Task.ID]
	if ok {
		persistedTask := m.TaskDb[te.Task.ID]
		if te.State == task.Completed && task.TaskFSM.ValidStateTransition(persistedTask.State, te.State) {
			m.stopTask(taskWorker, te.Task.ID.String())
			return
		}
		log.Printf("invalid request: existing task %s is in state %v and cannot transition to the completed state\n", persistedTask.ID.String(), persistedTask.State)
		return
	}

	w, err := m.SelectWorker(te.Task)
	if err != nil {
		log.Printf("Unable to find worker: %v.\n", err)
	}

	workerId := w.Name

	m.WorkerTaskMap[workerId][te.Task.ID] = true
	m.TaskWorkerMap[te.Task.ID] = workerId
	te.Task.State = task.Scheduled

	m.TaskDb[te.Task.ID] = &te.Task

	data, err := json.Marshal(te)
	if err != nil {
		log.Printf("Unable to marshal taskevent object: %v.\n", te)
	}

	url := fmt.Sprintf("http://%s/tasks", w.Ip)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error connecting to %v: %v\n", workerId, err)
		m.Pending.Enqueue(te)
		return
	}

	e := api.StandardResponse[task.Task]{}
	json.NewDecoder(resp.Body).Decode(&e)

	if resp.StatusCode != http.StatusCreated {
		log.Printf("Error: %s", e.ErrorMsg)
		return
	}

	var taskResult task.Task = e.Response
	log.Printf("%#v\n", taskResult)
}

func (m *Manager) UpdateTasks() {

	for _, workerString := range m.Workers {
		log.Printf("Checking worker %v for task updates", workerString)
		url := fmt.Sprintf("http://%s/tasks", workerString)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error connecting to %v: %v\n", workerString, err)
			return
		} else if resp.StatusCode != http.StatusOK {
			log.Printf("Error sending request: %v\n", err)
			return
		}

		d := json.NewDecoder(resp.Body)
		e := api.StandardResponse[[]task.Task]{}
		err = d.Decode(&e)
		if err != nil {
			log.Printf("Error unmarshalling tasks: %s\n", err.Error())
			return
		}

		for _, t := range e.Response {
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

func (m *Manager) UpdateTasksPeriodically() {

	ticker := time.NewTicker(time.Second * 12)
	for range ticker.C {
		m.UpdateTasks()
	}

}

func (m *Manager) GetTasks() []*task.Task {
	tasks := []*task.Task{}
	for _, v := range m.TaskDb {
		tasks = append(tasks, v)
	}
	return tasks
}

func (m *Manager) DoHealthChecksPeriodically() {
	ticker := time.NewTicker(time.Second * 18)
	for range ticker.C {

		log.Println("Performing task health check")
		m.doHealthChecks()
		log.Println("Task health checks completed")
	}
}

func (m *Manager) doHealthChecks() {
	for _, v := range m.TaskDb {
		if v.State == task.Dropped {
			continue
		}

		if v.State == task.Running && v.RestartCount < 3 {
			err := m.checkTaskHealth(*v)
			if err != nil {
				m.restartTask(v)
			}
		} else if v.State == task.Failed && v.RestartCount < 3 {
			m.restartTask(v)
		}
	}
}

func (m *Manager) checkTaskHealth(t task.Task) error {

	w := m.TaskWorkerMap[t.ID]
	hostPort := getHostPort(t.HostPorts)
	workerHost, _, _ := strings.Cut(w, ":")
	url := fmt.Sprintf("http://%s:%s%s", workerHost, *hostPort, t.HealthCheck)

	resp, err := http.Get(url)
	if err != nil {
		msg := fmt.Sprintf("Error connecting to health check %s", url)
		log.Println(msg)
		return errors.New(msg)
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Error health check for task %s did not return 200\n", t.ID)
		log.Println(msg)
		return errors.New(msg)
	}

	return nil
}

func (m *Manager) restartTask(t *task.Task) {
	w := m.TaskWorkerMap[t.ID]
	t.State = task.Scheduled
	t.RestartCount++
	m.TaskDb[t.ID] = t

	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Running,
		Timestamp: time.Now(),
		Task:      *t,
	}

	data, err := json.Marshal(te)
	if err != nil {
		log.Printf("Unable to marshal task object: %v.", t)
		return
	}

	url := fmt.Sprintf("http://%s/tasks", w)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error connecting to %v: %v", w, err)
		m.Pending.Enqueue(t)
		return
	}

	d := json.NewDecoder(resp.Body)
	e := api.StandardResponse[task.Task]{}
	err = d.Decode(&e)
	if err != nil {
		fmt.Printf("Error decoding response: %s\n", err.Error())
		return
	}

	if resp.StatusCode != http.StatusCreated {
		log.Printf("Response error (%d): %s", e.HttpStatusCode, e.ErrorMsg)
		return
	}

	log.Printf("%#v\n", e.Response)
}

func (m *Manager) stopTask(workerIp string, taskID string) {
	c := &http.Client{}

	url := fmt.Sprintf("http://%s/tasks/%s", workerIp, taskID)

	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	res, err := c.Do(req)

	if err != nil {
		log.Printf("Error sending request to worker: %v\n", err)
		return
	}

	if res.StatusCode != http.StatusNoContent {
		log.Printf("Error sending request: %v\n", err)
		return
	}

}

func getHostPort(ports nat.PortMap) *string {
	for k := range ports {
		return &ports[k][0].HostPort
	}
	return nil
}

func New(workers []string, scheduler scheduler.Scheduler) *Manager {

	taskDb := make(map[uuid.UUID]*task.Task)
	eventDb := make(map[uuid.UUID]*task.TaskEvent)
	workerTaskMap := make(map[string]map[uuid.UUID]interface{})
	taskWorkerMap := make(map[uuid.UUID]string)
	var nodes []*node.Node
	for worker := range workers {
		workerTaskMap[workers[worker]] = make(map[uuid.UUID]interface{})
		nAPI := fmt.Sprintf("http://%v", workers[worker])
		n := node.NewNode(workers[worker], nAPI, "worker", workers[worker])
		nodes = append(nodes, n)
	}

	return &Manager{
		Pending:       *queue.New(),
		TaskDb:        taskDb,
		EventDb:       eventDb,
		Workers:       workers,
		WorkerTaskMap: workerTaskMap,
		TaskWorkerMap: taskWorkerMap,
		WorkerNodes:   nodes,
		Scheduler:     scheduler,
	}
}
