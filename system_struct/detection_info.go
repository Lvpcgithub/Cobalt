package system_struct

type Task struct {
	SourceIP      string `json:"ip1"`
	DestinationIP string `json:"ip2"`
}
type ProbeResult struct {
	SourceIP      string  `json:"ip1"`
	DestinationIP string  `json:"ip2"`
	Delay         float32 `json:"tcp_delay"`
	Timestamp     string  `json:"timestamp"`
}
