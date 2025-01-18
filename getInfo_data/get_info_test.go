package getInfo_data

import (
	"Cobalt/dao"
	"log"
	"testing"
)

func TestFetchDeviceData(t *testing.T) {
	// 定义数据的来源 URL
	url := "http://localhost:8080/devices" // 替换为实际 API 地
	// 获取设备数据
	data, err := FetchDeviceData(url)
	if err != nil {
		log.Fatalf("Error fetching device data: %v", err)
	}
	db := dao.ConnectToDB()
	defer db.Close()
	// 处理设备数据
	ProcessDeviceData(db, data.Devices)
}
