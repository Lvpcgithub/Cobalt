package deviceApi

import (
	"log"
	"testing"
)

func TestDeviceInfo(t *testing.T) {
	r := DeviceInfo()
	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
