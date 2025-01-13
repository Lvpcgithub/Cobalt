package models

import (
	"Cobalt/system_struct"
	"Cobalt/topology_caculate"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"sync"
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
	//log.Printf("Stored data for %s->%s at %s", probeResult.SourceIP, probeResult.Delay, probeResult.Timestamp)
}

// RetrieveAndProcessData 从Redis中取出数据并进行计算 ****60s
// 为李雅普诺夫公式提供数据
var (
	Topology = system_struct.TopologyMatrix{
		Nodes: []string{},
		Links: make(map[string]system_struct.Link), // 初始化 map
	}
	Mu sync.Mutex // 用于并发安全
)

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
	values, err := redis.Values(conn.Do("LRANGE", key, -10, -1)) // 看一下逻辑
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
	}
	avgDelay := totalDelay / float64(len(values)) //除数是否是0
	stat, err := GetLatestCPUUsage(db, device_name2)
	if err != nil {
		log.Printf("Failed to get latest CPU usage for %s->%s: %v", ip1, ip2, err)
		return
	}
	aboveCpuMeans, belowCpuMeans, aboveCpuVars, belowCpuVars, err := GetCpuAvgAndVariance(db, 50, 50)
	if err != nil {
		return
	}
	net := system_struct.NetState{
		AboveThresholdCpuMeans: aboveCpuMeans,
		AboveThresholdCpuVars:  aboveCpuVars,
		BelowThresholdCpuMeans: belowCpuMeans,
		BelowThresholdCpuVars:  belowCpuVars,
	}
	node := system_struct.NodeState{
		CpuMean: stat.Mean,
		CpuVar:  stat.Variance,
	}
	//参数设置
	params := topology_caculate.SystemParams{
		ThresholdCpuMean: 50,
		ThresholdCpuVar:  50,
		Weight:           2,
	}
	//归一化数据计算
	normalCpuMean, normalCpuVar := params.Normalize(&node, &net)
	QMean, QVar, err := QueryVirtualQueueCPUByDeviceName(db, device_name2)
	if err != nil {
		log.Printf("Failed to query virtual CPU %v", err)
	}

	e := topology_caculate.Evaluate{
		Delay:         avgDelay,
		NormalCpuMean: normalCpuMean,
		NormalCpuVar:  normalCpuVar,
		QMean:         QMean,
		QVar:          QVar,
		Params:        params,
		State:         net,
	}
	VirtualQueueCPUMean := e.UpdateQMean()
	VirtualQueueCPUVariance := e.UpdateQVar()
	finalValue := e.DriftPlusPenalty()
	fmt.Println("finalValue:", finalValue)
	StoreToMySQL(db, device_name1, device_name2, avgDelay, stat.Mean, stat.Variance, VirtualQueueCPUMean, VirtualQueueCPUVariance)
	// 将结果加入到拓扑矩阵中
	Mu.Lock()
	k := fmt.Sprintf("%s:%s", ip1, ip2) // 构建键，例如 "192.168.1.1:192.168.1.2"

	// 插入或覆盖链接
	Topology.Links[k] = system_struct.Link{
		DeviceN1:   device_name1,
		DeviceN2:   device_name2,
		Source:     ip1,
		Target:     ip2,
		FinalValue: finalValue,
	}
	Mu.Unlock()

}

// 数据计算接收并存储到mysql
func StoreToMySQL(db *sql.DB, n1, n2 string, latency, mean, variance, VirtualQueueCPUMean, VirtualQueueCPUVariance float64) {
	// 创建 Link 结构体实例，准备插入的数据
	link := system_struct.Links{
		DeviceN1:                n1,
		DeviceN2:                n2,
		LinkLatency:             latency,
		N2CPUMean:               mean,
		N2CPUVariance:           variance,
		VirtualQueueCPUMean:     VirtualQueueCPUMean,     // 假设均值作为虚拟队列的 CPU 均值
		VirtualQueueCPUVariance: VirtualQueueCPUVariance, // 假设方差作为虚拟队列的 CPU 方差
	}
	// 调用 InsertLink 插入数据
	err := InsertLinks(db, link)
	if err != nil {
		fmt.Printf("Failed to insert link data: %v\n", err)
		return
	}
	fmt.Println("Link data inserted successfully!")
}
