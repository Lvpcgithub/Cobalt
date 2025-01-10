package models

import (
	"Cobalt/dao"
	"fmt"
	"log"
	"testing"
)

func TestRedis(t *testing.T) {

}

// 测试查询节点cpu和内存
func TestQueryDeviceIPs(t *testing.T) {
	db := dao.ConnectToDB()
	defer db.Close()
	// 查询设备和 IP
	devices, err := QueryDeviceIPs(db)
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}

	fmt.Println("Device Information:")
	for _, device := range devices {
		fmt.Printf("Device: %s, IPs: %s, Cores: %d, CPU: %s,MHZ: %f, Memory Total: %d, Memory Used: %d, Memory Used Percent: %.2f%%\n",
			device.DeviceName, device.IPs, device.CPUCores, device.CPUModelName, device.CPUMHz, device.MemoryTotal, device.MemoryUsed, device.MemoryUsedPercent)
	}
}
