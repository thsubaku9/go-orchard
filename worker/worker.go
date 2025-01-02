package worker

import (
	"errors"
	"fmt"
	"log"
	"orchard/task"
	"sync/atomic"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]*task.Task
	TaskCount atomic.Int32
}

func (w *Worker) CollectStats() {
	fmt.Println("I will collect stats")
}
func (w *Worker) RunTask() task.DockerResult {
	t := w.Queue.Dequeue()

	if t == nil {
		log.Printf("No task found")
		return task.DockerResult{Error: nil}
	}

	taskQueued, ok := t.(task.Task)

	if !ok {
		return task.DockerResult{Error: fmt.Errorf("%+v type casting error", t)}
	}

	taskPersisted := w.Db[taskQueued.ID]

	if taskPersisted == nil {
		taskPersisted = &taskQueued
		w.Db[taskQueued.ID] = &taskQueued
	}

	var result task.DockerResult

	if task.TaskFSM.ValidStateTransition(taskPersisted.State, taskQueued.State) {
		switch taskQueued.State {
		case task.Scheduled:
			result = w.StartTask(taskQueued)
		case task.Completed:
			result = w.StopTask(taskQueued)
		default:
			result.Error = errors.New("we should not get here")
		}
	} else {
		result.Error = fmt.Errorf("invalid transition from %v to %v", taskPersisted.State, taskQueued.State)
	}

	return result
}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}

func (w *Worker) StartTask(t task.Task) task.DockerResult {
	t.StartTime = time.Now().UTC()

	d, err := task.NewDocker(task.NewConfig(&t))

	if err != nil {
		log.Printf("Error creating client to stop task")
	}

	res := d.Run()

	if res.Error != nil {
		log.Printf("Err running task %v: %v\n", t.ID, res.Error)
		t.State = task.Failed
	} else {
		t.ContainerId = res.ContainerId
		t.State = task.Running
	}

	w.Db[t.ID] = &t
	return res

}

func (w *Worker) StopTask(t task.Task) task.DockerResult {
	d, err := task.NewDocker(task.NewConfig(&t))

	if err != nil {
		log.Printf("Error creating client to stop task")
	}

	res := d.Stop(t.ContainerId)

	if res.Error != nil {
		log.Printf("Error stopping container %v: %v\n", t.ContainerId, res.Error)
	}

	t.FinishTime = time.Now().UTC()
	t.State = task.Completed
	w.Db[t.ID] = &t

	log.Printf("Stopped and removed container %v for task %v\n", t.ContainerId, t.ID)
	return res
}
