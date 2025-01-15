package scheduler

import (
	"orchard/node"
	"orchard/task"
)

type Scheduler interface {
	SelectCandidateNodes(t task.Task, allNodes []*node.Node) []*node.Node
	ScoreNodes(t task.Task, candidateNodes []*node.Node) map[string]float64
	PickNode(scores map[string]float64, candidateNodes []*node.Node) *node.Node
}

type RoundRobin struct {
	Name       string
	LastWorker int
}

func (rr *RoundRobin) SelectCandidateNodes(t task.Task, allNodes []*node.Node) []*node.Node {
	return allNodes
}

func (rr *RoundRobin) ScoreNodes(t task.Task, candidateNodes []*node.Node) map[string]float64 {
	panic("not implemented") // TODO: Implement
}

func (rr *RoundRobin) PickNode(scores map[string]float64, candidateNodes []*node.Node) *node.Node {
	panic("not implemented") // TODO: Implement
}
