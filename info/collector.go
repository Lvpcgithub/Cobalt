package info

import (
	"github.com/shirou/gopsutil/v3/cpu"  // 获取CPU信息和使用率
	"github.com/shirou/gopsutil/v3/disk" // 获取磁盘信息，如分区和使用情况
	"github.com/shirou/gopsutil/v3/host" // 获取主机信息，如操作系统、平台等
	"github.com/shirou/gopsutil/v3/load" // 获取系统平均负载信息
	"github.com/shirou/gopsutil/v3/mem"  // 获取内存信息，如总量、使用量等
	"github.com/shirou/gopsutil/v3/net"  // 获取网络接口的I/O统计信息
	"io/ioutil"                          // 读取数据流
	"log"                                // 日志记录
	"net/http"                           // 执行HTTP请求
	"strings"
)

// GetIP 获取公网IP地址
// 通过向 http://icanhazip.com 发送GET请求来获取服务器的公网IP
func GetIP() string {
	resp, err := http.Get("http://icanhazip.com")
	if err != nil {
		log.Fatalf("Failed to get public IP: %v", err)
	}
	defer resp.Body.Close()

	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	return strings.TrimSpace(string(ip))
}

// GetCPUInfo 获取整体CPU信息及使用率
// 返回一个CPUInfo结构体，包含整体CPU的信息和使用率
func GetCPUInfo() CPUInfo {
	// 获取CPU的基本信息
	infos, err := cpu.Info()
	if err != nil {
		// 如果获取失败，记录错误并终止程序
		log.Fatalf("Failed to get CPU Info: %v", err)
	}

	// 获取每个CPU核心的使用率
	usage, err := cpu.Percent(0, true)
	if err != nil {
		// 如果获取失败，记录错误并终止程序
		log.Fatalf("Failed to get CPU usage: %v", err)
	}

	// 汇总CPU信息
	var totalCores int32
	var modelName string
	var mhz float64
	var cacheSize int32

	if len(infos) > 0 {
		totalCores = infos[0].Cores
		modelName = infos[0].ModelName
		mhz = infos[0].Mhz
		cacheSize = infos[0].CacheSize
	}

	// 计算整体使用率（取所有核心使用率的平均值）
	var totalUsage float64
	for _, u := range usage {
		totalUsage += u
	}
	averageUsage := totalUsage / float64(len(usage))

	// 创建并返回整体CPU信息
	return CPUInfo{
		Cores:     totalCores,
		ModelName: modelName,
		Mhz:       mhz,
		CacheSize: cacheSize,
		Usage:     averageUsage,
	}
}

// GetMemoryInfo 获取内存信息
// 返回一个MemoryInfo结构体，包含内存的总量、可用量、使用量等
func GetMemoryInfo() MemoryInfo {
	// 获取内存的使用情况
	v, err := mem.VirtualMemory()
	if err != nil {
		// 如果获取失败，记录错误并终止程序
		log.Fatalf("Failed to get memory info: %v", err)
	}

	// 返回内存信息的结构体
	return MemoryInfo{
		Total:       v.Total,
		Available:   v.Available,
		Used:        v.Used,
		UsedPercent: v.UsedPercent,
	}
}

// GetDiskInfo 获取整体磁盘信息
// 返回一个DiskInfo结构体，包含整个磁盘的使用情况
func GetDiskInfo() DiskInfo {
	// 获取根分区（"/"）的使用情况，这通常代表整个系统的磁盘使用情况
	usage, err := disk.Usage("/")
	if err != nil {
		// 如果获取失败，记录错误并终止程序
		log.Fatalf("Failed to get disk usage: %v", err)
	}

	// 返回整体磁盘信息
	return DiskInfo{
		Device:      "Overall",
		Total:       usage.Total,
		Free:        usage.Free,
		Used:        usage.Used,
		UsedPercent: usage.UsedPercent,
	}
}

// GetNetworkInfo 获取eth0接口的I/O统计信息
// 返回一个NetworkInfo结构体，包含eth0接口的流量统计信息
func GetNetworkInfo() NetworkInfo {
	// 获取所有网络接口的I/O统计信息
	interfaces, err := net.IOCounters(true)
	if err != nil {
		log.Fatalf("Failed to get network interfaces: %v", err)
	}

	for _, iface := range interfaces {
		if iface.Name == "eth0" {
			return NetworkInfo{
				InterfaceName: iface.Name,
				BytesSent:     iface.BytesSent,
				BytesRecv:     iface.BytesRecv,
				PacketsSent:   iface.PacketsSent,
				PacketsRecv:   iface.PacketsRecv,
			}
		}
	}

	// 如果没有找到 eth0 接口，返回一个零值的 NetworkInfo 结构体
	return NetworkInfo{}
}

// GetHostInfo 获取主机信息
// 返回一个HostInfo结构体，包含主机的名称、操作系统、平台版本等信息
func GetHostInfo() HostInfo {
	// 获取主机的基本信息
	info, err := host.Info()
	if err != nil {
		// 如果获取失败，记录错误并终止程序
		log.Fatalf("Failed to get host info: %v", err)
	}

	// 返回主机信息的结构体
	return HostInfo{
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		Uptime:          info.Uptime,
	}
}

// GetLoadInfo 获取系统平均负载信息
// 返回一个LoadInfo结构体，包含系统在过去1分钟、5分钟和15分钟内的平均负载
func GetLoadInfo() LoadInfo {
	// 获取系统平均负载信息
	avg, err := load.Avg()
	if err != nil {
		// 如果获取失败，记录错误并终止程序
		log.Fatalf("Failed to get system load: %v", err)
	}

	// 返回平均负载信息的结构体
	return LoadInfo{
		Load1:  avg.Load1,
		Load5:  avg.Load5,
		Load15: avg.Load15,
	}
}

// CollectSystemInfo 收集所有系统信息并返回
// 通过调用上述所有函数收集系统信息，并将它们组合成一个InfoData结构体返回
func CollectSystemInfo() InfoData {
	return InfoData{
		IP:          GetIP(),          // 获取公网IP
		CPUInfo:     GetCPUInfo(),     // 获取CPU信息和使用率
		MemoryInfo:  GetMemoryInfo(),  // 获取内存信息
		DiskInfo:    GetDiskInfo(),    // 获取磁盘信息
		NetworkInfo: GetNetworkInfo(), // 获取网络信息
		HostInfo:    GetHostInfo(),    // 获取主机信息
		LoadInfo:    GetLoadInfo(),    // 获取系统平均负载信息
	}
}
