package models

import (
	"Cobalt/system_struct"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
)

// 接受转化的数据，存储到redis
func CollectAndStoreData(conn redis.Conn, probeResult system_struct.ProbeResult) {
	key := fmt.Sprintf("probe:{%s:%s}", probeResult.SourceIP, probeResult.DestinationIP)
	data, err := json.Marshal(probeResult) // 将数据序列化为JSON格式
	if err != nil {
		log.Fatalf("Failed to marshal probe result: %v", err)
	}

	// 将序列化的数据存入Redis的列表
	_, err = conn.Do("RPUSH", key, data)
	if err != nil {
		log.Fatalf("Failed to store system info in Redis: %v", err)
	}
	log.Printf("Stored data for %s->%s at %s", probeResult.SourceIP, probeResult.Delay, probeResult.Timestamp)
}

// RetrieveAndProcessData 从Redis中取出数据并进行计算 ****60s
func RetrieveAndProcessData(conn redis.Conn, db *sql.DB, ip1 string, ip2 string) {
	device_name1, err := GetDeviceNameByIP(db, ip1)
	device_name2, err := GetDeviceNameByIP(db, ip2)
	if err != nil {
		log.Printf("Failed to get device name for %s->%s: %v", ip1, ip2, err)
		return
	}
	var totalDelay float64
	totalDelay = 0
	key := fmt.Sprintf("probe:{%s:%s}", ip1, ip2)
	fmt.Println("key:", key)
	// 获取最新10数据
	values, err := redis.Values(conn.Do("LRANGE", key, -10, -1))
	if err != nil {
		log.Fatalf("Failed to retrieve data from Redis: %v", err)
	}
	for _, v := range values {
		var probeResult system_struct.ProbeResult
		err := json.Unmarshal(v.([]byte), &probeResult)
		if err != nil {
			log.Printf("Failed to unmarshal probe result: %v", err)
			continue
		}
		totalDelay += probeResult.Delay
		//fmt.Println("计算数据,例如延迟：", probeResult.Delay)
	}
	avgDelay := totalDelay / float64(len(values))
	//fmt.Println(avgDelay)
	stat, err := GetLatestCPUUsage(db, device_name2)
	if err != nil {
		log.Printf("Failed to get latest CPU usage for %s->%s: %v", ip1, ip2, err)
		return
	}
	StoreToMySQL(db, device_name1, device_name2, avgDelay, stat.Mean, stat.Variance)
}

// 数据计算接收并存储到mysql
func StoreToMySQL(db *sql.DB, n1, n2 string, latency, mean, variance float64) {
	// 创建 Link 结构体实例，准备插入的数据
	link := system_struct.Links{
		DeviceN1:                n1,
		DeviceN2:                n2,
		LinkLatency:             latency,
		N2CPUMean:               mean,
		N2CPUVariance:           variance,
		VirtualQueueCPUMean:     0, // 假设均值作为虚拟队列的 CPU 均值
		VirtualQueueCPUVariance: 0, // 假设方差作为虚拟队列的 CPU 方差
	}
	// 调用 InsertLink 插入数据
	err := InsertLinks(db, link)
	if err != nil {
		fmt.Printf("Failed to insert link data: %v\n", err)
		return
	}
	fmt.Println("Link data inserted successfully!")
}
