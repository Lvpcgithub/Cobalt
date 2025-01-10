package getInfo_data

import (
	"log"
	"testing"
)

func TestFetchDeviceData(t *testing.T) {
	// 定义数据的来源 URL
	url := "http://localhost:8080/devices" // 替换为实际 API 地址

	// 获取设备数据
	data, err := FetchDeviceData(url)
	if err != nil {
		log.Fatalf("Error fetching device data: %v", err)
	}

	// 处理设备数据
	ProcessDeviceData(data.Devices)
}
