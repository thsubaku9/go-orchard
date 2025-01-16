package node

type Node struct {
	Name            string
	Ip              string
	Api             string
	Cores           int
	Memory          int
	MemoryAllocated int
	Disk            int
	DiskAllocated   int
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
