package main

import (
	"fmt"
	"log"
	"orchard/api"
	"orchard/manager"
	"orchard/scheduler"
	"orchard/task"
	"orchard/worker"
	"os"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/ttacon/chalk"
)

var red func(string) string = chalk.Red.NewStyle().WithTextStyle(chalk.Bold).Style
var green func(string) string = chalk.Green.NewStyle().WithTextStyle(chalk.Italic).Style

func createContainer() (*task.Docker, *task.DockerResult) {
	c := task.Config{
		Name:  "test-container-1",
		Image: "docker.io/library/alpine:latest",
		Env: []string{
			"SAMPLE_USER=cube",
		},
	}

	d, err := task.NewDocker(c)

	if err != nil {
		log.Printf("Error creating client %v\n", err)
		return nil, nil
	}

	result := d.Run()
	if result.Error != nil {
		log.Printf("Error running container : %v\n", result.Error)
		return nil, nil
	}
	log.Printf("Container %s is running with config %v\n", result.ContainerId, c)
	return d, &result
}

func purgeContainer(d *task.Docker, containerId string) *task.DockerResult {
	res := d.Stop(containerId)

	if res.Error != nil {
		log.Println(red(fmt.Sprintf("Error stopping container : %v\n", res.Error)))
		return nil
	}

	log.Println(green(fmt.Sprintf("Container %s has been stopped and removed", res.ContainerId)))
	return &res
}

func set_docker_envars() {
	if os.Getenv("DOCKER_HOST") == "" {
		os.Setenv("DOCKER_HOST", "unix:///Users/kernel/.docker/run/docker.sock")

	}
	if os.Getenv("DOCKER_API_VERSION") == "" {
		os.Setenv("DOCKER_API_VERSION", "1.45")
	}
}
func docker_main() {
	log.Printf("Container being created\n")
	dockerTask, dockerResult := createContainer()

	if dockerResult.Error != nil {
		log.Println(red(fmt.Sprintf("Create container has err : %v", dockerResult.Error)))
		os.Exit(1)
	}
	time.Sleep(time.Second * 8)

	fmt.Printf("Stopping container %s\n", dockerResult.ContainerId)
	_ = purgeContainer(dockerTask, dockerResult.ContainerId)
}

func create_task_config() task.Task {
	return task.Task{
		ID:    uuid.New(),
		Name:  "test-container-1",
		State: task.Scheduled,
		Image: "strm/helloworld-http",
		TaskConfig: task.Config{
			Name:  "test-container-1",
			Image: "docker.io/strm/helloworld-http:latest",
			Env:   []string{},
		},
	}
}

func worker_standalone_main() {
	w := worker.Worker{
		Name:  "Sample worker",
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	t := create_task_config()
	fmt.Println("starting task")
	w.AddTask(t)
	result := w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}

	t.ContainerId = result.ContainerId
	fmt.Printf("task %s is running in container %s\n", t.ID, t.ContainerId)
	fmt.Println("Sleepy time")
	time.Sleep(time.Second * 30)

	fmt.Printf("stopping task %s\n", t.ID)
	t.State = task.Completed
	w.AddTask(t)
	result = w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}
}

func worker_api_spinup(addr string, port string) {
	w := &worker.Worker{
		Name:  "Sample worker",
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	worker_api := worker.HttpApiWorker{
		HttpApi: api.HttpApi[worker.Worker]{
			Address: addr,
			Port:    port,
			Ref:     w,
		},
	}

	go w.RunTaskPeriodically()
	go w.CollectStats()
	go w.UpdateTasksPeriodically()
	go worker_api.StartServer()
}

func main() {

	set_docker_envars()
	worker_api_spinup("127.0.0.1", "7812")

	workers := []string{fmt.Sprintf("%s:%s", "127.0.0.1", "7812")}
	m := manager.New(workers, &scheduler.RoundRobin{})

	manager_api := manager.HttpApiManager{
		HttpApi: api.HttpApi[manager.Manager]{
			Address: "127.0.0.1",
			Port:    "9300",
			Ref:     m,
		},
	}

	for i := 0; i < 1; i++ {
		te := task.TaskEvent{
			ID:    uuid.New(),
			State: task.Pending,
			Task:  create_task_config(),
		}

		m.AddTask(te)
		m.SendWork()
	}

	go m.UpdateTasksPeriodically()
	go m.DoHealthChecksPeriodically()

	manager_api.StartServer()
}
