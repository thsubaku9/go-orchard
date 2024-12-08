package task

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
)

type Docker struct {
	Client *client.Client
	Config Config
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(ctx, d.Config.Image, types.ImagePullOptions{})

	if err != nil {
		log.Printf("Error pulling image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	io.Copy(os.Stdout, reader)

	return DockerResult{} // todok
}

func (d *Docker) Inspect(containerId string) DockerInspectResponse {
	dc, err := client.NewEnvClient()
	if err != nil {
		return DockerInspectResponse{Error: err}
	}

	ctx := context.Background()

	res, err := dc.ContainerInspect(ctx, containerId)

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
