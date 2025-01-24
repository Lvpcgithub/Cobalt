package getInfo_data

import (
	"Cobalt/kspTest"
	"Cobalt/models"
	"Cobalt/system_struct"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Response 定义整体 JSON 响应结构
type Response struct {
	Devices []system_struct.DeviceUseInfo `json:"devices"`
}

// fetchDeviceData 从指定 URL 获取设备数据
func FetchDeviceData(url string) (*Response, error) {
	// 发送 HTTP GET 请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from URL: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// 解析 JSON 数据
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return &response, nil
}

// PathAllocation 用于保存路径及其对应的资源分配数
type PathAllocation struct {
	PathIndex int      // 路径索引
	Path      []string // 存储路径的实际信息
	Allocated float64  // 分配的资源数
}

// ProcessDeviceData 处理设备数据并返回每条路径及其需求分配
func ProcessDeviceData(db *sql.DB, devices []system_struct.DeviceUseInfo, totalResource int, source, destination string) []PathAllocation {
	paths := kspTest.GetPaths(3, 2, 4, 1, source, destination)
	var avgCapacities []int              // 存储每条路径的 avgCapacity
	var pathAllocations []PathAllocation // 用于存储每条路径的需求分配数据

	// 遍历所有路径，计算每条路径的 avgCapacity
	for i, path := range paths {
		if len(path) > 2 {
			// 去除第一个和最后一个元素，保留中间节点
			trimmedPath := path[1 : len(path)-1]
			var totalCapacity = 0
			// 遍历中间节点
			for _, ip := range trimmedPath {
				// 查找对应的设备并计算节点能力
				for _, device := range devices {
					if ip == device.IPs {
						nodeCapacity, err := models.GetCapacityByCpuAndMemory(db, device.CPUUsage, device.MemoryUsedPercent)
						if err != nil {
							log.Printf("Error fetching capacity for device %s: %v\n", device.DeviceName, err)
							continue
						}
						totalCapacity += nodeCapacity.ConnectionsPerCycle
					}
				}
			}
			// 计算当前路径的平均能力
			avgCapacity := totalCapacity / len(trimmedPath)
			avgCapacities = append(avgCapacities, avgCapacity)
			fmt.Printf("Path %d: Avg Capacity = %d\n", i+1, avgCapacity)
		}
	}

	// 计算最大值和最小值，用于归一化
	var maxCapacity, minCapacity int
	for _, capacity := range avgCapacities {
		if capacity > maxCapacity {
			maxCapacity = capacity
		}
		if capacity < minCapacity || minCapacity == 0 {
			minCapacity = capacity
		}
	}

	// 归一化每条路径的 avgCapacity
	var normalizedCapacities []float64
	var sum float64
	for _, capacity := range avgCapacities {
		normalized := float64(capacity-minCapacity) / float64(maxCapacity-minCapacity)
		normalizedCapacities = append(normalizedCapacities, normalized)
		sum += normalized // 计算归一化后的比例总和
	}

	// 如果比例的总和不为 0，调整比例，使它们的总和为 1
	if sum > 0 {
		for i := range normalizedCapacities {
			normalizedCapacities[i] /= sum // 调整比例，使得它们的总和为 1
		}
	}
	// 计算并分配每条路径的资源
	for i, ratio := range normalizedCapacities {
		resourceForPath := float64(totalResource) * ratio // 根据比例分配资源
		pathAllocations = append(pathAllocations, PathAllocation{
			PathIndex: i + 1,
			Path:      paths[i], // 存储路径信息
			Allocated: resourceForPath,
		})
		fmt.Printf("Path %d gets %.2f units of resource\n", i+1, resourceForPath)
	}

	// 返回每条路径的分配需求
	return pathAllocations
}
