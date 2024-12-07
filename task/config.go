package task

import "github.com/docker/go-connections/nat"

type RestartPolicy string
type Config struct {
	Name          string
	AttachStdin   bool
	AttachStdout  bool
	AttachStderr  bool
	ExposedPorts  nat.PortSet
	Cmd           []string
	Image         string
	Cpu           float64
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy RestartPolicy
}

const (
	No             RestartPolicy = "no"
	Always         RestartPolicy = "always"
	UNLESS_STOPPED RestartPolicy = "unless-stopped"
	ON_FAILURE     RestartPolicy = "on-failure"
)
