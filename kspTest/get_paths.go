package kspTest

import (
	"fmt"
	"log"
)

// 获取路径数组，参数，迭代次数，k条路径，跳数限制，惩罚系数，源和目标
func GetPaths(MaxIterations, k, hopThreshold int, theta float64, source, destination string) [][]string {
	url := "http://127.0.0.1:8081/topology" // 替换为你的接口 URL
	data, err := fetchLinkInfo(url)
	if err != nil {
		log.Fatalf("Failed to fetch link info: %v", err)
	}

	// 将获取到的数据转换为矩阵
	matrix, ipList, reverseIpMap, err := jsonToMatrix(data)
	if err != nil {
		log.Fatalf("Failed to convert JSON to matrix: %v", err)
	}
	net := Network{
		Nodes: make([]Node, len(ipList)), // 节点数量与 IP 列表长度一致
		Links: matrix,                    // 使用转换后的矩阵作为 Links
	}

	// 存储不重复的路径
	var allPaths []Path
	flow := Flow{Source: 0, Destination: 2}
	graph := NewGraph(len(ipList))
	paths := KShortest(net, flow, k, hopThreshold, theta)

	// 根据每条路径更新容量
	for _, path := range paths {
		graph.UpdateCapacity(path)
	}
	iterations := 0
	// 计算最大流
	s, sink := 0, 2
	for len(allPaths) < k && iterations < MaxIterations {
		fmt.Println("K Shortest Paths:")
		for i, p := range paths {
			fmt.Printf("Path %d: Nodes: %v, Latency: %d\n", i+1, p.Nodes, p.Latency)
		}
		maxFlow := graph.EdmondsKarp(s, sink)
		fmt.Printf("Maximum Flow: %d\n", maxFlow)

		// 输出流量矩阵，查看哪些链路成为瓶颈
		fmt.Println("Flow matrix:")
		for i := 0; i < graph.Nodes; i++ {
			fmt.Println(graph.Flow[i])
		}

		// 查找最小割
		minCut := graph.findMinCut(s)
		fmt.Println("Minimum Cut Edges:")
		for _, edge := range minCut {
			fmt.Printf("Edge from %d to %d with capacity %d\n", edge.From, edge.To, edge.Capacity)
		}

		// 查找最小割中容量最大的边
		maxCapacityEdge := FindMaxCapacityEdgeInMinCut(minCut)
		fmt.Printf("Max Edge from %d to %d with capacity %d\n", maxCapacityEdge.From, maxCapacityEdge.To, maxCapacityEdge.Capacity)

		// 存储符合条件的路径
		var validPaths []Path

		// 遍历当前路径列表
		for _, path := range paths {
			// 判断路径是否包含最大容量的最小割边
			if containsMaxCapacityMinCutEdge(path, maxCapacityEdge) {
				// 如果路径包含最大容量的最小割边，保留延迟最小的一条路径
				if len(validPaths) == 0 {
					validPaths = append(validPaths, path)
				} else {
					// 保留延迟最小的一条路径
					if path.Latency < validPaths[0].Latency {
						validPaths = []Path{path}
					}
				}
			} else {
				// 如果路径不包含该边，则保留路径，首先检查是否已经存在
				if !containsPath(allPaths, path) {
					//fmt.Println("合格的：", path)
					allPaths = append(allPaths, path)
				}
			}
		}

		// 打印有效路径
		for i, p := range validPaths {
			if !containsPath(allPaths, p) { // 确保路径不重复
				allPaths = append(allPaths, p)
			}
			fmt.Printf("Path %d: Nodes: %v, Latency: %d\n", i+1, p.Nodes, p.Latency)
		}

		// 更新拓扑图：删除最小割中的最大容量边
		net.RemoveEdge(maxCapacityEdge.From, maxCapacityEdge.To)

		// 重新计算 K 最短路径
		paths = KShortest(net, flow, k, 4, 1)
		// 根据每条路径更新容量
		graph.ClearFlowMatrix()
		for _, path := range paths {
			graph.UpdateCapacity(path)
		}

		fmt.Println("Flow matrix after updating capacity:", paths)
		for i := 0; i < graph.Nodes; i++ {
			fmt.Println(graph.Flow[i])
		}
		iterations++
	}

	// 打印所有找到的不重复路径并转换为 IP 地址
	fmt.Println("Final K Shortest Paths (no duplicates):")
	var ipPaths [][]string
	for i, p := range allPaths {
		// 将路径节点索引转换为 IP 地址
		ipPath := make([]string, len(p.Nodes))
		for idx, node := range p.Nodes {
			ipPath[idx] = reverseIpMap[node] // 使用 reverseIpMap 转换索引为 IP 地址
		}
		// 输出 IP 地址路径
		fmt.Printf("Path %d: Nodes: %v, Latency: %d\n", i+1, ipPath, p.Latency)
		ipPaths = append(ipPaths, ipPath)
	}

	// 返回 IP 地址路径的二维数组
	return ipPaths
}
