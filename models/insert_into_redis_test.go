package models

import (
	"Cobalt/dao"
	"log"
	"sync"
	"testing"
)

func TestRetrieveAndProcessData(t *testing.T) {
	conn := dao.ConnRedis()
	defer conn.Close()
	db := dao.ConnectToDB()
	defer db.Close()
	ipAddresses := []string{"192.168.0.1", "192.168.1.1", "192.168.2.1"}
	var wg sync.WaitGroup

	for i := 0; i < len(ipAddresses); i++ {
		for j := 0; j < len(ipAddresses); j++ {
			if i != j {
				wg.Add(1)
				go func(ip1, ip2 string) {
					defer wg.Done()
					RetrieveAndProcessData(conn, db, ip1, ip2)
				}(ipAddresses[i], ipAddresses[j])
			}
		}
	}

	// 等待所有 goroutine 完成
	wg.Wait()

}

func TestTopologyInfo(t *testing.T) {
	r := TopologyInfo()
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
