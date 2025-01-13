package getInfo_data

import (
	"Cobalt/system_struct"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// processDeviceData 处理设备数据
func ProcessDeviceData(devices []system_struct.DeviceUseInfo) {
	for _, device := range devices {
		fmt.Printf("Device Name: %s\n", device.DeviceName)
		fmt.Printf("IPs: %s\n", device.IPs)
		fmt.Printf("CPU Cores: %d\n", device.CPUCores)
		fmt.Printf("CPU Model Name: %s\n", device.CPUModelName)
		fmt.Printf("CPU MHz: %.2f\n", device.CPUMHz)
		fmt.Printf("Memory Total: %d\n", device.MemoryTotal)
		fmt.Printf("Memory Used: %d\n", device.MemoryUsed)
		fmt.Printf("Memory Used Percent: %.2f%%\n", device.MemoryUsedPercent)
		fmt.Println("-----------------------------")
	}
}
