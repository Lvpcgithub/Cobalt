package kspTest

// the numbers of nodes in a network start from 0, e.g., 10 nodes are 0-9.
type Network struct {
	Nodes []Node  `json:"nodes"`
	Links [][]int `json:"links"` // latency of every link, -1 means that there is no link between two nodes
}

type Node struct {
}

type Flow struct {
	Source      int `json:"source"`
	Destination int `json:"destination"`
}

type Path struct {
	Nodes   []int `json:"nodes"`   // nodes on a path
	Latency int   `json:"latency"` // total latency of a path
}

type RoutingResult struct {
	Net   Network `json:"net"`
	Flows []Flow  `json:"flows"`
}

// maxflow使用的
// Edge 结构表示一条边
type Edge struct {
	From, To, Capacity int
}

// Graph 结构表示一个网络图
type Graph struct {
	Capacity [][]int // 容量矩阵
	Flow     [][]int // 流量矩阵
	Nodes    int     // 节点数量
}
