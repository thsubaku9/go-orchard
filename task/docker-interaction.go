package task

import (
	"context"
	"io"
	"log"
	"os"
	"sync"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
)

type Docker struct {
	Client *client.Client
	Config Config
}

var clientPool sync.Pool = sync.Pool{
	New: func() any {
		client, _ := client.NewEnvClient()
		return &Docker{
			Client: client,
		}
	},
}

func NewClientFromPool() *Docker {
	return clientPool.Get().(*Docker)
}

func NewDocker(c Config) (*Docker, error) {
	dockerClient, err := client.NewEnvClient()

	if err != nil {
		log.Printf("Error creating client %v\n", err)
		return nil, err
	}

	return &Docker{
		Client: dockerClient,
		Config: c,
	}, nil
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(ctx, d.Config.Image, types.ImagePullOptions{})

	if err != nil {
		log.Printf("Error pulling image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)

	cc := container.Config{
		Image:        d.Config.Image,
		Tty:          false,
		Env:          d.Config.Env,
		ExposedPorts: d.Config.ExposedPorts,
		Cmd:          []string{"sh"},
	}

	hc := container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: string(d.Config.RestartPolicy)},
		Resources: container.Resources{
			Memory:    d.Config.Memory,
			CPUShares: int64(d.Config.Cpu),
		},
		PublishAllPorts: true,
	}

	res, err := d.Client.ContainerCreate(ctx, &cc, &hc, nil, d.Config.Name)
	if err != nil {
		log.Printf("Error creating container using image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerStart(ctx, res.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Printf("Error starting container %s: %v\n", res.ID, err)
		return DockerResult{Error: err}
	}

	out, err := d.Client.ContainerLogs(ctx, res.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})

	if err != nil {
		log.Printf("Error getting logs for container %s: %v\n", res.ID, err)
		return DockerResult{Error: err}
	}

	defer out.Close()
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return DockerResult{ContainerId: res.ID, Action: "start", Result: "success"}
}

func (d *Docker) Stop(containerId string) DockerResult {
	log.Printf("Attempting to stop container %v", containerId)
	ctx := context.Background()
	err := d.Client.ContainerStop(ctx, containerId, nil)
	if err != nil {
		log.Printf("Error stopping container %s: %v\n", containerId, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         false,
	})

	if err != nil {
		log.Printf("Error removing container %s: %v\n", containerId, err)
		return DockerResult{Error: err}
	}

	return DockerResult{Action: "stop", Result: "success", Error: nil}
}

func (d *Docker) Inspect(containerId string) DockerInspectResponse {
	dc := NewClientFromPool().Client
	res, err := dc.ContainerInspect(context.Background(), containerId)

	if err != nil {
		return DockerInspectResponse{Error: err}
	}

	return DockerInspectResponse{Container: &res}
}

type DockerResult struct {
	Error       error
	Action      string
	ContainerId string
	Result      string
}

type DockerInspectResponse struct {
	Error     error
	Container *types.ContainerJSON
}
