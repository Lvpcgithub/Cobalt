package models

import (
	"Cobalt/dao"
	"Cobalt/system_struct"
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
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
	//RetrieveAndProcessData(conn, probeResult.SourceIP, probeResult.DestinationIP)
}

// RetrieveAndProcessData 从Redis中取出数据并进行计算 ****60s
func RetrieveAndProcessData(conn redis.Conn, ip1 string, ip2 string) {
	var totalDelay float32
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
		fmt.Println("计算数据,例如延迟：", probeResult.Delay)
		// 进行必要的计算, 比如时延分析或其他
		//ProcessAndStoreToMySQL(probeResult)
	}
	avgDelay := totalDelay / float32(len(values))
	fmt.Println(avgDelay)
	StoreToMySQL(ip1, ip2, avgDelay, time.Now().Format("2006-01-02 15:04:05"))
}

// 数据计算接收并存储到mysql
func StoreToMySQL(sourceIP string, destinationIP string, delay float32, timestamp string) {
	fmt.Println(sourceIP, destinationIP, delay, timestamp)
	db := dao.ConnectToDB()
	err := InsertLinkInfo(db, sourceIP, destinationIP, delay, timestamp)
	if err != nil {
		return
	}

}
