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

type RoundRobin struct {
	LastWorker int
}

func (rr *RoundRobin) SelectCandidateNodes(t task.Task, allNodes []*node.Node) []*node.Node {
	return allNodes
}

func (rr *RoundRobin) ScoreNodes(t task.Task, candidateNodes []*node.Node) map[string]float64 {

	nodeScores := make(map[string]float64)
	var newWorker int = rr.LastWorker
	rr.LastWorker++
	rr.LastWorker %= len(candidateNodes)

	for idx, node := range candidateNodes {
		nodeScores[node.Name] = 1.0

		if idx == newWorker {
			nodeScores[node.Name] = 0.1
		}

	}

	return nodeScores
}

func (rr *RoundRobin) PickNode(scores map[string]float64, candidateNodes []*node.Node) *node.Node {
	var bestNode *node.Node
	var lowestScore float64
	for idx, v := range candidateNodes {
		if idx == 0 {
			bestNode = v
			lowestScore = scores[v.Name]
			continue
		}

		if scores[v.Name] < lowestScore {
			bestNode = v
			lowestScore = scores[v.Name]
		}
	}

	return bestNode
}

func (rr *RoundRobin) Name() string {
	return "round-robin"
}
