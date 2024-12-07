package scheduler

type Scheduler interface {
	SelectCandidateNodes()
	ScoreNodes()
	PickNode()
}
