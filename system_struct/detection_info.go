package system_struct

type Task struct {
	SourceIP      string `json:"ip1"`
	DestinationIP string `json:"ip2"`
}
type ProbeResult struct {
	SourceIP      string  `json:"ip1"`
	DestinationIP string  `json:"ip2"`
	Delay         float64 `json:"tcp_delay"`
	Timestamp     string  `json:"timestamp"`
}

// Links 是用于存储 links 表的数据结构
type Links struct {
	DeviceN1                string  `json:"device_n1"`
	DeviceN2                string  `json:"device_n2"`
	LinkLatency             float64 `json:"link_latency"`
	N2CPUMean               float64 `json:"n2_cpu_mean"`
	N2CPUVariance           float64 `json:"n2_cpu_variance"`
	VirtualQueueCPUMean     float64 `json:"virtual_queue_cpu_mean"`
	VirtualQueueCPUVariance float64 `json:"virtual_queue_cpu_variance"`
}
