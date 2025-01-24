package kspTest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Link 数据结构定义
type Link struct {
	DeviceN1   string  `json:"device_n1"`
	DeviceN2   string  `json:"device_n2"`
	IP1        string  `json:"ip1"`
	IP2        string  `json:"ip2"`
	FinalValue float64 `json:"final_value"` // 保持为float64，方便解析
}

// Data 结构体定义
type Data struct {
	Nodes []interface{}   `json:"nodes"`
	Links map[string]Link `json:"links"`
}

// 从 URL 获取数据
func fetchLinkInfo(url string) (*Data, error) {
	// 发送请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching data from URL: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// 打印出来查看获取的原始数据
	fmt.Println("Raw Data:", string(body))

	// 定义接收的数据结构
	var result Data
	// 解析 JSON 数据
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return &result, nil
}

// 将数据转化为矩阵
func jsonToMatrix(data *Data) ([][]int, []string, map[int]string, error) {
	// 提取所有 IP 地址并创建映射
	ipMap := make(map[string]int)
	var ipList []string
	count := 0

	// 遍历所有链接，提取 IP 地址并建立映射
	for _, link := range data.Links {
		// 对于 IP1，若不存在于 ipMap 中则添加
		if _, exists := ipMap[link.IP1]; !exists {
			ipMap[link.IP1] = count
			ipList = append(ipList, link.IP1)
			count++
		}

		// 对于 IP2，若不存在于 ipMap 中则添加
		if _, exists := ipMap[link.IP2]; !exists {
			ipMap[link.IP2] = count
			ipList = append(ipList, link.IP2)
			count++
		}
	}

	// 初始化一个空矩阵，大小为节点数
	matrixSize := len(ipList)
	matrix := make([][]int, matrixSize)
	for i := range matrix {
		matrix[i] = make([]int, matrixSize)
	}

	// 填充矩阵
	for _, link := range data.Links {
		i := ipMap[link.IP1] // 获取 IP1 对应的节点索引
		j := ipMap[link.IP2] // 获取 IP2 对应的节点索引

		// 填充矩阵：从 IP1 到 IP2 的延迟
		matrix[i][j] = int(link.FinalValue)
	}

	// 创建反向 IP 映射
	reverseIpMap := make(map[int]string)
	for index, ip := range ipList {
		reverseIpMap[index] = ip
	}

	return matrix, ipList, reverseIpMap, nil
}
