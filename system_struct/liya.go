package system_struct

// 节点状态
type NodeState struct {
	CpuMean float64 // 当前CPU均值
	CpuVar  float64 // 当前CPU方差
}

// 网络拓扑状态
type NetState struct {
	AboveThresholdCpuMeans []float64 // 所有超过阈值节点CPU均值，升序排序
	BelowThresholdCpuMeans []float64 // 所有未超过阈值节点CPU均值，升序排序
	AboveThresholdCpuVars  []float64 // 所有超过阈值节点CPU方差，升序排序
	BelowThresholdCpuVars  []float64 // 所有未超过阈值节点CPU方差，升序排序
}
