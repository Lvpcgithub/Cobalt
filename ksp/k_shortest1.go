package test_info

import "container/heap"

type Path struct {
	Cost  float64 // 当前路径的总权重
	Node  int     // 当前节点
	Route []int   // 当前路径的节点列表
}

type PriorityQueue []*Path

// 实现 heap.Interface 接口
func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Cost < pq[j].Cost
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*Path))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// 检查节点是否已经在路径中，避免形成环
func contains(route []int, node int) bool {
	for _, n := range route {
		if n == node {
			return true
		}
	}
	return false
}

// KShortestPaths1 计算 K 条最短路径并对路径跳数超过三跳的路径应用惩罚
func KShortestPaths1(matrix [][]int, source, target, k int, theta float64, skip int) [][2]interface{} {
	n := len(matrix) // 节点总数
	pq := &PriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &Path{Cost: 0, Node: source, Route: []int{source}})
	shortestPaths := [][2]interface{}{} // 存储路径和对应的惩罚后的权重值
	for pq.Len() > 0 && len(shortestPaths) < k {
		current := heap.Pop(pq).(*Path)
		// 如果当前节点是目标节点，将路径和权重加入结果集
		if current.Node == target {
			// 计算路径惩罚
			penalizedCost := current.Cost
			if len(current.Route) > skip {
				// 路径跳数超过 3 时，应用惩罚
				penalizedCost += (float64(len(current.Route)-skip) * theta) * current.Cost
			}

			// 将路径和惩罚后的权重加入结果集
			//if !isPathConverged(current.Route, shortestPaths) {
			shortestPaths = append(shortestPaths, [2]interface{}{current.Route, penalizedCost})

			//}
			continue
		}
		// 遍历当前节点的邻居节点
		for neighbor := 0; neighbor < n; neighbor++ {
			if matrix[current.Node][neighbor] != MAX_INT { // 只考虑有效路径
				// 检查环
				if contains(current.Route, neighbor) {
					continue // 跳过带环的路径
				}
				// 构建新的路径
				newCost := current.Cost + float64(matrix[current.Node][neighbor])
				newRoute := append([]int{}, current.Route...)
				newRoute = append(newRoute, neighbor)
				heap.Push(pq, &Path{Cost: newCost, Node: neighbor, Route: newRoute})
			}
		}
	}
	return shortestPaths
}
