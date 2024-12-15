package main

import (
	"fmt"
	"log"
	"orchard/task"
	"os"
	"time"

	"github.com/docker/engine-api/client"
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

	dc, err := client.NewEnvClient()

	if err != nil {
		log.Printf("Error creating client %v\n", err)
		return nil, nil
	}

	d := task.Docker{
		Client: dc,
		Config: c}

	result := d.Run()
	if result.Error != nil {
		log.Printf("Error running container : %v\n", result.Error)
		return nil, nil
	}
	log.Printf("Container %s is running with config %v\n", result.ContainerId, c)
	return &d, &result
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

func main() {

	log.Printf("Container being created\n")
	if os.Getenv("DOCKER_HOST") == "" {
		os.Setenv("DOCKER_HOST", "unix:///Users/kernel/.docker/run/docker.sock")

	} else if os.Getenv("DOCKER_API_VERSION") == "" {
		os.Setenv("DOCKER_API_VERSION", "1.45")
	}

	dockerTask, dockerResult := createContainer()

	if dockerResult.Error != nil {
		log.Println(red(fmt.Sprintf("Create container has err : %v", dockerResult.Error)))
		os.Exit(1)
	}
	time.Sleep(time.Second * 8)

	fmt.Printf("Stopping container %s\n", dockerResult.ContainerId)
	_ = purgeContainer(dockerTask, dockerResult.ContainerId)
}
