package kspTest

import (
	"fmt"
	"testing"
)

func TestPrintKShortestPaths(t *testing.T) {
	net := Network{
		Nodes: []Node{{}, {}, {}, {}}, // 4 nodes
		Links: [][]int{
			{0, 1, 1, 2}, // Latencies from node 0
			{1, 0, 1, 1}, // Latencies from node 1
			{1, 1, 0, 2}, // Latencies from node 2
			{2, 1, 2, 0}, // Latencies from node 3
		},
	}

	flow := Flow{Source: 0, Destination: 3}
	k := 4

	// 初始化图，假设有 5 个节点
	graph := NewGraph(4)
	paths := KShortest(net, flow, k, 4, 1)

	// 根据每条路径更新容量
	for _, path := range paths {
		graph.UpdateCapacity(path)
	}
	// 计算最大流
	source, sink := 0, 3
	maxFlow := graph.EdmondsKarp(source, sink)
	fmt.Printf("Maximum Flow: %d\n", maxFlow)

	// 输出流量矩阵，查看哪些链路成为瓶颈
	fmt.Println("Flow matrix:")
	for i := 0; i < graph.Nodes; i++ {
		fmt.Println(graph.Flow[i])
	}
	// 运行 K 最短路径算法

	// 打印 K 条最短路径
	for i, p := range paths {
		fmt.Printf("Path %d: Nodes: %v, Latency: %d\n", i+1, p.Nodes, p.Latency)
	}
}
func TestPrintKShortestPaths1(t *testing.T) {
	/*
		1、首先计算出 K 最短路径paths，
		2、将路径中不包含容量最大的那个最小割的边保留下来，包含最小割边的路径保留下来延迟最小的一条，其余路径删除
		3、将符合条件的路径加入到allpath里
		4、更新拓扑图：删除最小割中容量大的边
		5、再次计算最短路径，最大流、最小割，并继续处理剩余的路径，直到找到k条的不重复路径。
	*/
	// 网络拓扑，包含6个节点
	net := Network{
		Nodes: []Node{{}, {}, {}, {}, {}, {}}, // 6 nodes
		Links: [][]int{
			//   0   1   2   3   4   5
			{0, 2, -1, -1, 1, -1},   // Node 0
			{-1, 0, 2, -1, -1, -1},  // Node 1
			{-1, -1, 0, 1, 3, -1},   // Node 2
			{-1, -1, -1, 0, 1, 2},   // Node 3
			{-1, -1, -1, -1, 0, 2},  // Node 4
			{-1, -1, -1, -1, -1, 0}, // Node 5
		},
	}
	// 存储不重复的路径
	var allPaths []Path
	iterations := 0
	flow := Flow{Source: 0, Destination: 5}
	k := 4
	graph := NewGraph(6)
	paths := KShortest(net, flow, k, 4, 1)

	// 根据每条路径更新容量
	for _, path := range paths {
		graph.UpdateCapacity(path)
	}

	// 计算最大流
	source, sink := 0, 5
	for len(allPaths) < k && iterations < 3 {
		fmt.Println("K Shortest Paths:")
		for i, p := range paths {
			fmt.Printf("Path %d: Nodes: %v, Latency: %d\n", i+1, p.Nodes, p.Latency)
		}
		maxFlow := graph.EdmondsKarp(source, sink)
		fmt.Printf("Maximum Flow: %d\n", maxFlow)

		// 输出流量矩阵，查看哪些链路成为瓶颈
		fmt.Println("Flow matrix:")
		for i := 0; i < graph.Nodes; i++ {
			fmt.Println(graph.Flow[i])
		}

		// 查找最小割
		minCut := graph.findMinCut(source)
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

	// 打印所有找到的不重复路径
	fmt.Println("Final K Shortest Paths (no duplicates):")
	for i, p := range allPaths {
		fmt.Printf("Path %d: Nodes: %v, Latency: %d\n", i+1, p.Nodes, p.Latency)
	}
}
func TestPrintKShortestPaths2(t *testing.T) {
	// 网络拓扑，包含6个节点
	/*
		1、首先计算出 K 最短路径paths，
		2、将路径中不包含容量最大的那个最小割的边保留下来，包含最小割边的路径保留下来延迟最小的一条，其余路径删除
		3、将符合条件的路径加入到allpath里
		4、更新拓扑图：删除最小割中容量大的边
		5、再次计算最短路径，最大流、最小割，并继续处理剩余的路径，直到找到k条的不重复路径。
	*/
	/*
		之前逻辑保持不变，在allpath继续添加如下三条逻辑
			1、allpath是按照latency从小到大排序的 √
			2、多次迭代过程之中，如果迭代过后有其他path比allpath中的更优，那么需要插入新的path，×
			2、根据ksp特性，最开始算出的k条路径具有最小延迟，只需要和allpaths最后一条路径比较
			3、迭代完之后去allpath中的前k跳路径输出
	*/
	net := Network{
		Nodes: []Node{{}, {}, {}, {}, {}, {}}, // 6 nodes
		Links: [][]int{
			//   0   1   2   3   4   5
			{0, 3, -1, 5, 7, -1},  // Node 0
			{3, 0, 2, -1, 4, -1},  // Node 1
			{-1, 2, 0, 4, -1, -1}, // Node 2
			{5, -1, 4, 0, 3, 8},   // Node 3
			{7, 4, -1, 3, 0, 6},   // Node 4
			{-1, -1, -1, 8, 6, 0}, // Node 5
		},
	}

	// 存储不重复的路径
	var allPaths []Path
	iterations := 0
	flow := Flow{Source: 0, Destination: 5}
	k := 6
	graph := NewGraph(6)
	paths := KShortest(net, flow, k, 4, 1)

	// 根据每条路径更新容量
	for _, path := range paths {
		graph.UpdateCapacity(path)
	}

	// 计算最大流
	source, sink := 0, 5
	//len(allPaths) < k &&
	for iterations < 3 {
		fmt.Println("K Shortest Paths:")
		for i, p := range paths {
			fmt.Printf("Path %d: Nodes: %v, Latency: %d\n", i+1, p.Nodes, p.Latency)
		}
		maxFlow := graph.EdmondsKarp(source, sink)
		fmt.Printf("Maximum Flow: %d\n", maxFlow)

		// 输出流量矩阵，查看哪些链路成为瓶颈
		fmt.Println("Flow matrix:")
		for i := 0; i < graph.Nodes; i++ {
			fmt.Println(graph.Flow[i])
		}

		// 查找最小割
		minCut := graph.findMinCut(source)
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

		// 排序 allPaths
		sortPathsByLatency(allPaths)
		fmt.Println("allPaths (no duplicates):")
		for i, p := range allPaths {
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

	// 打印所有找到的不重复路径
	fmt.Println("Final K Shortest Paths (no duplicates):")
	for i, p := range allPaths {
		fmt.Printf("Path %d: Nodes: %v, Latency: %d\n", i+1, p.Nodes, p.Latency)
	}
}
