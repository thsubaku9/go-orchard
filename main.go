package main

import (
	"fmt"
	"log"
	"orchard/task"
	"os"

	"github.com/docker/engine-api/client"
	"github.com/ttacon/chalk"
)

var red func(string) string = chalk.Red.NewStyle().WithBackground(chalk.Black).WithTextStyle(chalk.Bold).Style

func createContainer() (*task.Docker, *task.DockerResult) {
	c := task.Config{
		Name:  "test-container-1",
		Image: "postgres:13",
		Env: []string{
			"POSTGRES_USER=cube",
			"POSTGRES_PASSWORD=secret",
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

	log.Println(red(fmt.Sprintf("Container %s has been stopped and removed\n", res.ContainerId)))
	return &res
}

func main() {

	log.Printf("Container being created\n")

	dockerTask, dockerResult := createContainer()

	if dockerResult.Error != nil {
		log.Println(red(fmt.Sprintf("Create container has err : %v", dockerResult.Error)))
		os.Exit(1)
	}

	var res string
	fmt.Println("Press anything to stop container")
	fmt.Scan(res)

	fmt.Printf("stopping container %s\n", dockerResult.ContainerId)
	_ = purgeContainer(dockerTask, dockerResult.ContainerId)
}
