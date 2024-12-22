package models

import (
	"Cobalt/system_struct"
	"database/sql"
	"fmt"
	"time"
)

// 插入节点信息
func InsertDeviceInfo(db *sql.DB, info system_struct.DeviceInfo) error {
	query := `
		INSERT INTO device_info (
			device_name, 
			cpu_cores, cpu_model_name, cpu_mhz, cpu_cache_size, cpu_usage,
			memory_total, memory_available, memory_used, memory_used_percent,
			disk_device, disk_total, disk_free, disk_used, disk_used_percent,
			network_interface_name, network_bytes_sent, network_bytes_recv,
			network_packets_sent, network_packets_recv,
			hostname, os, platform, platform_version, uptime,
			load1, load5, load15, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("Timestamp: %s\n", timestamp)

	_, err := db.Exec(query,
		info.DeviceName,
		info.CPUInfo.Cores, info.CPUInfo.ModelName, info.CPUInfo.Mhz, info.CPUInfo.CacheSize, info.CPUInfo.Usage, //5
		info.MemoryInfo.Total, info.MemoryInfo.Available, info.MemoryInfo.Used, info.MemoryInfo.UsedPercent, //4
		info.DiskInfo.Device, info.DiskInfo.Total, info.DiskInfo.Free, info.DiskInfo.Used, info.DiskInfo.UsedPercent, //5
		info.NetworkInfo.InterfaceName, info.NetworkInfo.BytesSent, info.NetworkInfo.BytesRecv, //3
		info.NetworkInfo.PacketsSent, info.NetworkInfo.PacketsRecv, //2
		info.HostInfo.Hostname, info.HostInfo.OS, info.HostInfo.Platform, info.HostInfo.PlatformVersion, info.HostInfo.Uptime, //5
		info.LoadInfo.Load1, info.LoadInfo.Load5, info.LoadInfo.Load15, timestamp, //4
	)
	return err
}

func InsertIpInfo(db *sql.DB, info system_struct.IpList) error {
	query := `
		INSERT INTO ip_list (device_name, ip_address, ip_type, description)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, info.DeviceName, info.IpAddress, info.IpType, info.Description)
	return err
}

// 插入链路信息
func InsertLinks(db *sql.DB, link system_struct.Links) error {
	// 插入 SQL 语句
	query := `
		INSERT INTO links (
			device_n1, device_n2, link_latency, 
			n2_cpu_mean, n2_cpu_variance, 
			virtual_queue_cpu_mean, virtual_queue_cpu_variance
		) VALUES (?, ?, ?, ?, ?, ?, ?);
	`

	// 执行插入操作
	_, err := db.Exec(query, link.DeviceN1, link.DeviceN2, link.LinkLatency,
		link.N2CPUMean, link.N2CPUVariance, link.VirtualQueueCPUMean, link.VirtualQueueCPUVariance)
	if err != nil {
		return fmt.Errorf("failed to insert link data: %v", err)
	}

	// 成功插入后返回 nil
	return nil
}
