package test_info

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const MAX_INT = 1<<31 - 1 // 定义为很大的值（最大整数）
// 从url中获取信息
func fetchLinkInfo(url string) ([]LinkInfo, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var links []LinkInfo
	if err := json.Unmarshal(body, &links); err != nil {
		return nil, err
	}
	return links, nil
}

// 将数据转化为矩阵
func jsonToMatrix(links []LinkInfo) ([][]int, map[string]int) {

	nodeMap := make(map[string]int)
	index := 0
	for _, link := range links {
		if _, exists := nodeMap[link.IP1]; !exists {
			nodeMap[link.IP1] = index
			index++
		}
		if _, exists := nodeMap[link.IP2]; !exists {
			nodeMap[link.IP2] = index
			index++
		}
	}

	// 初始化矩阵并设置默认值为 MAX_INT
	n := len(nodeMap)
	matrix := make([][]int, n)
	for i := range matrix {
		matrix[i] = make([]int, n)
		for j := range matrix[i] {
			matrix[i][j] = MAX_INT // 默认值为最大值
		}
	}

	// 填充矩阵的已有链接值
	for _, link := range links {
		from := nodeMap[link.IP1]
		to := nodeMap[link.IP2]
		matrix[from][to] = link.LinkWeight
	}

	return matrix, nodeMap
}

func PrintInfo(s string) {
	url := "http://127.0.0.1:8080/" + s
	links, err := fetchLinkInfo(url)
	if err != nil {
		log.Fatal("Error fetching link info:", err)
	}

	matrix, nodeMap := jsonToMatrix(links)

	// 打印邻接矩阵
	fmt.Println("Adjacency Matrix:")
	for _, row := range matrix {
		for _, val := range row {
			if val == MAX_INT {
				// 如果是最大值，输出 "INF"
				fmt.Printf("%4s ", "INF")
			} else {
				// 格式化输出，确保对齐
				fmt.Printf("%4d ", val)
			}
		}
		fmt.Println()
	}
	// 定义源和目标节点以及路径数量 k
	source := nodeMap["192.168.32.1"]
	target := nodeMap["192.168.32.5"]
	k := 6
	// 迭代查找 k 条不汇聚的路径
	theta := 0.1 // 惩罚系数
	skip := 3    // 路径跳数超过 3 的惩罚
	finalPaths := FindKNonConvergingPaths(matrix, source, target, k, theta, skip)

	fmt.Println("\nFinal K Non-Converging Paths:")
	for i, path := range finalPaths {
		fmt.Printf("Path %d: %v, Cost: %v\n", i+1, path[0], path[1])
	}
}

const maxConvergeIterations = 4 // 最大收敛次数

// 计算 K 条不汇聚的最短路径
func FindKNonConvergingPaths(matrix [][]int, source, target, k int, theta float64, skip int) [][2]interface{} {
	var finalPaths [][2]interface{}       // 存储不汇聚的路径
	visitedPaths := make(map[string]bool) // 用于存储已生成的路径
	convergeCount := 0
	for len(finalPaths) < k {
		// 1. 调用 KShortestPaths1 查找 K 条最短路径
		shortestPaths := KShortestPaths1(matrix, source, target, k, theta, skip)
		fmt.Println("\n  K Paths:")
		for i, path := range shortestPaths {
			fmt.Printf("Path %d: %v, Cost: %v\n", i+1, path[0], path[1])
		}
		// 如果没有找到新的路径，终止迭代
		if len(shortestPaths) == 0 {
			break
		}
		convergeCount++
		if convergeCount >= maxConvergeIterations {
			fmt.Println("No new paths found for multiple iterations, exiting.")
			break
		}
		// 2. 更新路径和汇聚检测
		// 3. 使用 updateMatrixForConvergedPaths 排除路径汇聚的节点
		matrix = updateMatrixForConvergedPaths(matrix, shortestPaths)
		fmt.Println("更新 Matrix:")
		for _, row := range matrix {
			for _, val := range row {
				if val == MAX_INT {
					// 如果是最大值，输出 "INF"
					fmt.Printf("%4s ", "INF")
				} else {
					// 格式化输出，确保对齐
					fmt.Printf("%4d ", val)
				}
			}
			fmt.Println()
		}
		// 4. 从 shortestPaths 中筛选出不汇聚的路径，且确保路径唯一
		for _, path := range shortestPaths {
			// 将路径转换为字符串以便在 visitedPaths 中查找
			pathStr := pathToString(path[0].([]int))
			// 如果路径已经存在 finalPaths 中，则跳过添加
			if visitedPaths[pathStr] || isPathConverged(path[0].([]int), finalPaths) {
				continue
			}
			// 标记路径为已访问
			visitedPaths[pathStr] = true
			// 将路径添加到 finalPaths
			finalPaths = append(finalPaths, path)
			// 如果已经添加了 k 条路径，提前退出
			if len(finalPaths) >= k {
				break
			}
		}
		// 如果最终路径已达 k 条，停止迭代
		if len(finalPaths) >= k {
			break
		}
	}
	return finalPaths
}

// 将路径转换为字符串，便于在 visitedPaths 中查找
func pathToString(path []int) string {
	return fmt.Sprintf("%v", path) // 将路径转换为字符串格式
}

// 检查路径是否与已选路径汇聚
func isPathConverged(path []int, finalPaths [][2]interface{}) bool {
	for _, finalPath := range finalPaths {
		route := finalPath[0].([]int)
		// 判断路径是否汇聚
		if x, y := pathsConverge(path, route); x != -1 || y != -1 {
			return true
		}
	}
	return false
}

// updateMatrixForConvergedPaths 用来更新矩阵，排除汇聚路径
func updateMatrixForConvergedPaths(matrix [][]int, allPaths [][2]interface{}) [][]int {
	newMatrix := make([][]int, len(matrix))
	for i := range matrix {
		newMatrix[i] = make([]int, len(matrix[i]))
		copy(newMatrix[i], matrix[i]) // 复制原始矩阵
	}
	fmt.Println("new Matrix:")
	for _, row := range matrix {
		for _, val := range row {
			if val == MAX_INT {
				// 如果是最大值，输出 "INF"
				fmt.Printf("%4s ", "INF")
			} else {
				// 格式化输出，确保对齐
				fmt.Printf("%4d ", val)
			}
		}
		fmt.Println()
	}
	// 遍历所有路径，找出哪些路径汇聚
	for i := 0; i < len(allPaths); i++ {
		// 排除一跳路径
		if len(allPaths[i][0].([]int)) == 2 {
			continue
		}
		for j := i + 1; j < len(allPaths); j++ {
			// 排除一跳路径
			if len(allPaths[j][0].([]int)) == 2 {
				continue
			}
			// 比较路径 i 和路径 j 是否汇聚
			x, y := pathsConverge(allPaths[i][0].([]int), allPaths[j][0].([]int))
			fmt.Printf("Updating matrix for converged paths between %v and %v at indices (%d, %d)\n", allPaths[i][0], allPaths[j][0], x, y)
			if x != -1 && y != -1 { // 只有在索引有效的情况下才更新矩阵
				// 更新矩阵，排除汇聚的路径
				newMatrix[allPaths[i][0].([]int)[x]][allPaths[i][0].([]int)[y]] = MAX_INT
			}
		}
	}
	return newMatrix // 返回修改后的矩阵
}

// pathsConverge 判断两条路径是否汇聚
func pathsConverge(path1, path2 []int) (int, int) {
	// 找到交点并判断是否有相同的节点对
	for i := 1; i < len(path1)-1; i++ { // 排除源节点和目标节点
		for j := 1; j < len(path2)-1; j++ {
			if path1[i] == path2[j] && path1[i+1] == path2[j+1] {
				// 找到路径汇聚，返回交点的索引
				return i, i + 1
			}
		}
	}
	return -1, -1 // 没有找到汇聚点
}
