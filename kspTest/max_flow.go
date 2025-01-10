package kspTest

import (
	"fmt"
	"math"
	"sort"
)

// 创建一个新的图
func NewGraph(nodes int) *Graph {
	graph := &Graph{
		Capacity: make([][]int, nodes),
		Flow:     make([][]int, nodes),
		Nodes:    nodes,
	}

	// 初始化容量矩阵和流量矩阵
	for i := 0; i < nodes; i++ {
		graph.Capacity[i] = make([]int, nodes)
		graph.Flow[i] = make([]int, nodes)
	}
	return graph
}

// Graph 结构体中更新容量的函数
func (g *Graph) ClearFlowMatrix() {
	for i := 0; i < g.Nodes; i++ {
		for j := 0; j < g.Nodes; j++ {
			g.Capacity[i][j] = 0
		}
	}
}

// UpdateCapacity 更新路径经过的链路的容量
func (g *Graph) UpdateCapacity(path Path) {
	// 每条路径经过的链路容量设为1
	for i := 0; i < len(path.Nodes)-1; i++ {
		u := path.Nodes[i]
		v := path.Nodes[i+1]
		g.Capacity[u][v] += 1 // 你可以根据需求调整容量
	}
}

// Edmonds-Karp 算法计算最大流
func (g *Graph) EdmondsKarp(source, sink int) int {
	maxFlow := 0
	for {
		// 查找增广路径
		parent := make([]int, g.Nodes)
		for i := 0; i < g.Nodes; i++ {
			parent[i] = -1
		}
		pathFlow := bfs(g, source, sink, parent)
		if pathFlow == 0 {
			break // 如果没有增广路径，结束
		}

		// 更新流量和残余容量
		maxFlow += pathFlow
		v := sink
		for v != source {
			u := parent[v]
			g.Flow[u][v] += pathFlow
			g.Flow[v][u] -= pathFlow // 更新反向边流量
			v = u
		}
	}
	return maxFlow
}

// 使用 BFS 查找增广路径
func bfs(g *Graph, source, sink int, parent []int) int {
	visited := make([]bool, g.Nodes)
	queue := []int{source}
	visited[source] = true

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]

		for v := 0; v < g.Nodes; v++ {
			// 如果有残余容量，并且未访问过
			if !visited[v] && g.Capacity[u][v]-g.Flow[u][v] > 0 {
				parent[v] = u
				if v == sink {
					// 找到增广路径，返回路径上的最小容量
					return minCapacity(g, parent, source, sink)
				}
				queue = append(queue, v)
				visited[v] = true
			}
		}
	}
	return 0
}

// 计算增广路径上的最小容量
func minCapacity(g *Graph, parent []int, source, sink int) int {
	pathFlow := math.MaxInt32
	v := sink
	for v != source {
		u := parent[v]
		pathFlow = min(pathFlow, g.Capacity[u][v]-g.Flow[u][v])
		v = u
	}
	return pathFlow
}

// 辅助函数：最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// findMinCut 找到最小割的边
func (g *Graph) findMinCut(source int) []Edge {
	// 找出从源点可达的所有节点
	reachable := make([]bool, g.Nodes)
	dfs(g, source, reachable)

	// 查找源点可达与不可达节点之间的边
	var minCut []Edge
	for u := 0; u < g.Nodes; u++ {
		if reachable[u] {
			for v := 0; v < g.Nodes; v++ {
				if !reachable[v] && g.Capacity[u][v] > 0 {
					minCut = append(minCut, Edge{u, v, g.Capacity[u][v]})
				}
			}
		}
	}
	return minCut
}

// DFS 深度优先搜索，标记源节点可达的所有节点
func dfs(g *Graph, u int, reachable []bool) {
	reachable[u] = true
	for v := 0; v < g.Nodes; v++ {
		if !reachable[v] && g.Capacity[u][v]-g.Flow[u][v] > 0 {
			dfs(g, v, reachable)
		}
	}
}

/*// 删除最小割中容量较大的边
func (net *Network) RemoveMinCutEdges(minCut []Edge) {
	// 如果最小割为空，直接返回
	if len(minCut) == 0 {
		return
	}

	// 找出最小割中容量最大的边
	maxCapacityEdge := minCut[0]
	for _, edge := range minCut {
		if edge.Capacity > maxCapacityEdge.Capacity {
			maxCapacityEdge = edge
		}
	}

	// 删除容量最大的边
	u, v := maxCapacityEdge.From, maxCapacityEdge.To
	fmt.Printf("Deleting edge from %d to %d with capacity %d\n", u, v, maxCapacityEdge.Capacity)
	net.Links[u][v] = -1 // 删除该边（设置为不可达）
	//net.Links[v][u] = -1 // 如果是无向图，反向边也需要删除
}*/
// FindMaxCapacityEdgeInMinCut 找到最小割中容量最大的边
func FindMaxCapacityEdgeInMinCut(minCut []Edge) Edge {
	// 如果最小割为空，返回一个空的边（容量为0）
	if len(minCut) == 0 {
		return Edge{}
	}

	// 找出最小割中容量最大的边
	maxCapacityEdge := minCut[0]
	for _, edge := range minCut {
		if edge.Capacity > maxCapacityEdge.Capacity {
			maxCapacityEdge = edge
		}
	}

	return maxCapacityEdge
}

// RemoveEdge 删除指定的边（将该边的容量设置为 -1，表示不可达）
func (net *Network) RemoveEdge(u, v int) {
	// 删除边
	fmt.Printf("Deleting edge from %d to %d\n", u, v)
	net.Links[u][v] = 50 // 设置该边为大的值（设置为不可达）
	// 如果是无向图，反向边也需要删除
	// net.Links[v][u] = -1 // 如果需要删除反向边，取消注释
}

// containsMaxCapacityMinCutEdge 判断路径中是否包含最小割中容量最大的那条边
func containsMaxCapacityMinCutEdge(path Path, maxCapacityEdge Edge) bool {
	// 遍历路径中的每一条边，检查路径中是否包含最大容量的最小割边
	for i := 0; i < len(path.Nodes)-1; i++ {
		from := path.Nodes[i]
		to := path.Nodes[i+1]

		// 检查当前路径是否包含最大容量的最小割边
		if from == maxCapacityEdge.From && to == maxCapacityEdge.To {
			// 如果路径包含该边，返回 true
			return true
		}
	}
	// 如果路径中不包含该边，返回 false
	return false
}

// 检查路径是否已存在
func containsPath(allPaths []Path, newPath Path) bool {
	for _, p := range allPaths {
		// 比较路径的节点是否完全相同
		if equalPaths(p, newPath) {
			return true
		}
	}
	return false
}

// 比较两个路径是否相同
func equalPaths(path1, path2 Path) bool {
	if len(path1.Nodes) != len(path2.Nodes) {
		return false
	}
	for i := range path1.Nodes {
		if path1.Nodes[i] != path2.Nodes[i] {
			return false
		}
	}
	return true
}

// 按延迟排序路径
func sortPathsByLatency(paths []Path) {
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Latency < paths[j].Latency
	})
}
