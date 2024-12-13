package system_struct

// 节点信息结构体
type CPUInfo struct {
	Cores     int32   `json:"cores"`
	ModelName string  `json:"model_name"`
	Mhz       float64 `json:"mhz"`
	CacheSize int32   `json:"cache_size"`
	Usage     float64 `json:"usage"`
}

type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

type DiskInfo struct {
	Device      string  `json:"device"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

type NetworkInfo struct {
	InterfaceName string `json:"interface_name"`
	BytesSent     uint64 `json:"bytes_sent"`
	BytesRecv     uint64 `json:"bytes_recv"`
	PacketsSent   uint64 `json:"packets_sent"`
	PacketsRecv   uint64 `json:"packets_recv"`
}

type HostInfo struct {
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	Uptime          uint64 `json:"uptime"`
}

type LoadInfo struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

type SystemInfo struct {
	IP          string      `json:"ip"`
	CPUInfo     CPUInfo     `json:"cpu_info"`
	MemoryInfo  MemoryInfo  `json:"memory_info"`
	DiskInfo    DiskInfo    `json:"disk_info"`
	NetworkInfo NetworkInfo `json:"network_info"`
	HostInfo    HostInfo    `json:"host_info"`
	LoadInfo    LoadInfo    `json:"load_info"`
}
