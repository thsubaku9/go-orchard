package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"orchard/api"
	"orchard/metrics"
)

type Node struct {
	Name            string
	Ip              string
	Api             string
	Cores           int
	Memory          int
	MemoryAllocated int
	Disk            int
	DiskAllocated   int
	Stats           metrics.Metrics
	Role            string
	TaskCount       int
}

func NewNode(name string, api string, role string, ip string) *Node {
	return &Node{
		Name: name,
		Ip:   ip,
		Api:  api,
		Role: role,
	}
}

func (n *Node) GetStats() (*metrics.Metrics, error) {
	url := fmt.Sprintf("%s/stats", n.Api)
	resp, err := http.Get(url)
	if err != nil {
		msg := fmt.Sprintf("Unable to connect to %v. Permanent failure.\n", n.Api)
		log.Println(msg)
		return nil, errors.New(msg)
	}

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("Error retrieving stats from %v: %v", n.Api, err)
		log.Println(msg)
		return nil, errors.New(msg)
	}

	var respBody api.StandardResponse[metrics.Metrics]
	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&respBody)
	if err != nil {
		msg := fmt.Sprintf("error decoding message while getting stats for node %s", n.Name)
		log.Println(msg)
		return nil, errors.New(msg)
	}

	n.Memory = int(respBody.Response.Memory.Total)
	n.Disk = int(respBody.Response.Disk.Total)
	n.Stats = respBody.Response

	return &n.Stats, nil
}
