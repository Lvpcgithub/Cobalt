package tcp_probe

import "time"

// ProbeTask 结构体定义
type ProbeTask struct {
	IP1  string `json:"ip1"`
	IP2  string `json:"ip2"`
	Port string `json:"port"`
}

// ProbeResult 结构体定义
type ProbeResult struct {
	IP1       string        `json:"ip1"`
	IP2       string        `json:"ip2"`
	TCPDelay  time.Duration `json:"tcp_delay"`
	Timestamp time.Time     `json:"timestamp"`
}
