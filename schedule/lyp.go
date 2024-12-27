package schedule

import (
	"math"
)

// 定义系统参数
type SystemParams struct {
	thresholdCpuMean float64 // CPU均值阈值
	thresholdCpuVar  float64 // CPU方差阈值
	weight           float64 // 时延的权重
}

// 选路评价
type Evaluate struct {
	delay         float64      // 当前链路时延
	normalCpuMean float64      // cpu均值归一化
	normalCpuVar  float64      // cpu方差归一化
	qMean         float64      // CPU均值虚拟队列
	qVar          float64      // CPU方差虚拟队列
	params        SystemParams // 系统参数
	state         NetState     // 网络拓扑
}

// 节点状态
type NodeState struct {
	cpuMean float64 // 当前CPU均值
	cpuVar  float64 // 当前CPU方差
}

// 网络拓扑状态
type NetState struct {
	aboveThresholdCpuMeans []float64 // 所有超过阈值节点CPU均值，升序排序
	belowThresholdCpuMeans []float64 // 所有未超过阈值节点CPU均值，升序排序
	aboveThresholdCpuVars  []float64 // 所有超过阈值节点CPU方差，升序排序
	belowThresholdCpuVars  []float64 // 所有未超过阈值节点CPU方差，升序排序
}

// 归一化
func (s *SystemParams) Normalize(node *NodeState, net *NetState) (float64, float64) {
	// 归一化结果
	var CpuMeanNormalization, CpuVarNormalization float64
	// 判断当前CPU均值是否超过阈值
	isAboveThresholdCpuMean := func(x float64) bool {
		return x > s.thresholdCpuMean
	}(node.cpuMean)

	if isAboveThresholdCpuMean {
		// 计算当前超过阈值的节点排序位置
		rank := 1
		for _, cpu := range net.aboveThresholdCpuMeans {
			if cpu < node.cpuMean {
				rank++
			}
		}
		// 分位数归一化
		CpuMeanNormalization = float64(rank-1) / float64(len(net.aboveThresholdCpuMeans)-1)
	} else {
		// 计算当前未超过阈值的节点排序位置
		rank := 1
		for _, cpu := range net.belowThresholdCpuMeans {
			if cpu < node.cpuMean {
				rank++
			}
		}
		// 分位数归一化
		CpuMeanNormalization = -(1 - float64(rank-1)/float64(len(net.belowThresholdCpuMeans)-1))
	}
	// 判断当前CPU方差是否超过阈值
	isAboveThresholdCpuVar := func(x float64) bool {
		return x > s.thresholdCpuVar
	}(node.cpuVar)

	if isAboveThresholdCpuVar {
		// 计算当前超过阈值的节点排序位置
		rank := 1
		for _, cpu := range net.aboveThresholdCpuVars {
			if cpu < node.cpuVar {
				rank++
			}
		}
		// 分位数归一化
		CpuVarNormalization = float64(rank-1) / float64(len(net.aboveThresholdCpuVars)-1)
	} else {
		// 计算当前未超过阈值的节点排序位置
		rank := 1
		for _, cpu := range net.belowThresholdCpuVars {
			if cpu < node.cpuVar {
				rank++
			}
		}
		// 分位数归一化
		CpuVarNormalization = -(1 - float64(rank-1)/float64(len(net.belowThresholdCpuVars)-1))
	}
	return CpuMeanNormalization, CpuVarNormalization
}

// 更新 CPU 均值虚拟队列
func (e *Evaluate) updateQMean() float64 {
	return math.Max(e.qMean+e.normalCpuMean, 0)
}

// 更新 CPU 方差虚拟队列
func (e *Evaluate) updateQVar() float64 {
	return math.Max(e.qVar+e.normalCpuVar, 0)
}

// 计算漂移加惩罚式子
func (e *Evaluate) driftPlusPenalty() float64 {
	delayPart := e.params.weight * e.delay
	meanPart := e.qMean * e.normalCpuMean
	varPart := e.qVar * e.normalCpuVar

	// 返回扩展公式总和
	return delayPart + meanPart + varPart
}
