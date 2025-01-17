package scheduler

import (
	"orchard/node"
	"orchard/task"
)

type Scheduler interface {
	SelectCandidateNodes(t task.Task, allNodes []*node.Node) []*node.Node
	ScoreNodes(t task.Task, candidateNodes []*node.Node) map[string]float64
	PickNode(scores map[string]float64, candidateNodes []*node.Node) *node.Node
	Name() string
}
