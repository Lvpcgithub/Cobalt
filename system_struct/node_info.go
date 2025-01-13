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

type DeviceInfo struct {
	DeviceName  string      `json:"device_name"`
	CPUInfo     CPUInfo     `json:"cpu_info"`
	MemoryInfo  MemoryInfo  `json:"memory_info"`
	DiskInfo    DiskInfo    `json:"disk_info"`
	NetworkInfo NetworkInfo `json:"network_info"`
	HostInfo    HostInfo    `json:"host_info"`
	LoadInfo    LoadInfo    `json:"load_info"`
}
type IpList struct {
	DeviceName  string `json:"device_name" db:"device_name"` // 外键引用设备表
	IpAddress   string `json:"ip_address" db:"ip_address"`   // IP地址（支持IPv4和IPv6）
	IpType      string `json:"ip_type" db:"ip_type"`         // IP类型 ('IPv4' 或 'IPv6')
	Description string `json:"description" db:"description"` // IP地址的备注信息
}

type CPUStats struct {
	DeviceName string  `json:"device_name"` // 设备名
	Mean       float64 `json:"mean"`        // CPU 使用率的均值
	Variance   float64 `json:"variance"`    // CPU 使用率的方差
}

// DeviceIP 结构体表示查询结果
type DeviceUseInfo struct {
	DeviceName        string  // 对应 device_name
	IPs               string  // 对应 ips
	CPUCores          int     // 对应 cpu_cores
	CPUModelName      string  // 对应 cpu_model_name
	CPUMHz            float64 // 对应 cpu_mhz
	MemoryTotal       int64   // 对应 memory_total
	MemoryUsed        int64   // 对应 memory_used
	MemoryUsedPercent float64 // 对应 memory_used_percent
}
type TopologyMatrix struct {
	Nodes []string        `json:"nodes"`
	Links map[string]Link `json:"links"` // 使用 map 存储链接
}

type Link struct {
	DeviceN1   string  `json:"device_n1"`
	DeviceN2   string  `json:"device_n2"`
	Source     string  `json:"ip1"`
	Target     string  `json:"ip2"`
	FinalValue float64 `json:"final_value"`
}
