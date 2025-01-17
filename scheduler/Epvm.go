package scheduler

import (
	"math"
	"orchard/node"
	"orchard/task"
	"time"
)

const LIEB float64 = 1.53960071783900203869

/*
Based on the https://mosix.cs.huji.ac.il/pub/ocja.pdf paper
*/
type Epvm struct {
}

func (epvm *Epvm) SelectCandidateNodes(t task.Task, allNodes []*node.Node) []*node.Node {
	var candidateNodes []*node.Node

	for _, applicantNode := range allNodes {
		if checkDisk(t, applicantNode) {
			candidateNodes = append(candidateNodes, applicantNode)
		}
	}

	return candidateNodes
}

func (epvm *Epvm) ScoreNodes(t task.Task, candidateNodes []*node.Node) map[string]float64 {

	nodeScores := make(map[string]float64)
	max_jobs := 4

	for _, node := range candidateNodes {
		cpuUsage := calculateCpuUsage(node)
		cpuLoad := cpuUsage / math.Pow(2, 0.8)

		memoryAllocated := float64(node.Stats.MemUsedKb()) + float64(node.MemoryAllocated)
		memoryAllocatedPercentage := memoryAllocated / float64(node.Memory)

		newMemPercent := (memoryAllocated + float64(t.Memory/1000)) / float64(node.Memory)
		memCost := (math.Pow(LIEB, newMemPercent) - math.Pow(LIEB, memoryAllocatedPercentage)) + (math.Pow(LIEB, (float64(node.TaskCount+1))/float64(max_jobs)) - math.Pow(LIEB, (float64(node.TaskCount))/float64(max_jobs)))
		cpuCost := (math.Pow(LIEB, cpuLoad) - math.Pow(LIEB, cpuLoad)) + (math.Pow(LIEB, (float64(node.TaskCount+1))/float64(max_jobs)) - math.Pow(LIEB, (float64(node.TaskCount))/float64(max_jobs)))

		nodeScores[node.Name] = memCost + cpuCost

	}

	return nodeScores
}

func calculateCpuUsage(node *node.Node) float64 {

	stat1, _ := node.GetStats()
	time.Sleep(3 * time.Second)
	stat2, _ := node.GetStats()
	stat1Idle := stat1.CPU.TimeStat.Idle + stat1.CPU.TimeStat.Iowait
	stat2Idle := stat2.CPU.TimeStat.Idle + stat2.CPU.TimeStat.Iowait

	stat1NonIdle := stat1.CPU.TimeStat.User + stat1.CPU.TimeStat.Nice + stat1.CPU.TimeStat.System + stat1.CPU.TimeStat.Irq + stat1.CPU.TimeStat.Softirq + stat1.CPU.TimeStat.Steal
	stat2NonIdle := stat2.CPU.TimeStat.User + stat2.CPU.TimeStat.Nice + stat2.CPU.TimeStat.System + stat2.CPU.TimeStat.Irq + stat2.CPU.TimeStat.Softirq + stat2.CPU.TimeStat.Steal

	stat1Total := stat1Idle + stat1NonIdle
	stat2Total := stat2Idle + stat2NonIdle
	total := stat2Total - stat1Total
	idle := stat2Idle - stat1Idle

	if total == 0 && idle == 0 {
		return 0.00
	} else {
		return (float64(total) - float64(idle)) / float64(total)
	}

}

func (epvm *Epvm) PickNode(scores map[string]float64, candidateNodes []*node.Node) *node.Node {
	minCost := 0.00
	var bestNode *node.Node
	for idx, node := range candidateNodes {
		if idx == 0 {
			minCost = scores[node.Name]
			bestNode = node
			continue
		}
		if scores[node.Name] < minCost {
			minCost = scores[node.Name]
			bestNode = node
		}
	}
	return bestNode
}

func (epvm *Epvm) Name() string {
	return "EPVM"
}

func checkDisk(t task.Task, applicantNode *node.Node) bool {
	return t.Disk <= applicantNode.Disk-applicantNode.DiskAllocated
}
