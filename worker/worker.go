package worker

import (
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
func (w *Worker) RunTask() {
	fmt.Println("I will start or stop a task")
}

func (w *Worker) StartTask() {
	fmt.Println("I will start a task")
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
