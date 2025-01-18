package topology_caculate

import (
	"Cobalt/system_struct"
	"math"
)

// 选路评价
type Evaluate struct {
	Delay         float64                // 当前链路时延
	NormalCpuMean float64                // cpu均值归一化
	NormalCpuVar  float64                // cpu方差归一化
	QMean         float64                // CPU均值虚拟队列
	QVar          float64                // CPU方差虚拟队列
	Params        SystemParams           // 系统参数
	State         system_struct.NetState // 网络拓扑
}

// 定义系统参数
type SystemParams struct {
	ThresholdCpuMean float64 // CPU均值阈值
	ThresholdCpuVar  float64 // CPU方差阈值
	Weight           float64 // 时延的权重
}

// 归一化
func (s *SystemParams) Normalize(node *system_struct.NodeState, net *system_struct.NetState) (float64, float64) {
	// 归一化结果
	var CpuMeanNormalization, CpuVarNormalization float64

	// 判断当前CPU均值是否超过阈值
	isAboveThresholdCpuMean := func(x float64) bool {
		return x > s.ThresholdCpuMean
	}(node.CpuMean)

	if isAboveThresholdCpuMean {
		// 计算当前超过阈值的节点排序位置
		rank := 1
		for _, cpu := range net.AboveThresholdCpuMeans {
			if cpu < node.CpuMean {
				rank++
			}
		}

		// 检查除数是否为零
		lenAbove := len(net.AboveThresholdCpuMeans)
		if lenAbove > 1 { // 防止除数为零
			CpuMeanNormalization = float64(rank-1) / float64(lenAbove-1)
		} else {
			CpuMeanNormalization = 0 // 如果只有一个元素，归一化结果为0
		}
	} else {
		// 计算当前未超过阈值的节点排序位置
		rank := 1
		for _, cpu := range net.BelowThresholdCpuMeans {
			if cpu < node.CpuMean {
				rank++
			}
		}

		// 检查除数是否为零
		lenBelow := len(net.BelowThresholdCpuMeans)
		if lenBelow > 1 { // 防止除数为零
			CpuMeanNormalization = -(1 - float64(rank-1)/float64(lenBelow-1))
		} else {
			CpuMeanNormalization = 0 // 如果只有一个元素，归一化结果为0
		}
	}

	// 判断当前CPU方差是否超过阈值
	isAboveThresholdCpuVar := func(x float64) bool {
		return x > s.ThresholdCpuVar
	}(node.CpuVar)

	if isAboveThresholdCpuVar {
		// 计算当前超过阈值的节点排序位置
		rank := 1
		for _, cpu := range net.AboveThresholdCpuVars {
			if cpu < node.CpuVar {
				rank++
			}
		}

		// 检查除数是否为零
		lenAbove := len(net.AboveThresholdCpuVars)
		if lenAbove > 1 { // 防止除数为零
			CpuVarNormalization = float64(rank-1) / float64(lenAbove-1)
		} else {
			CpuVarNormalization = 0 // 如果只有一个元素，归一化结果为0
		}
	} else {
		// 计算当前未超过阈值的节点排序位置
		rank := 1
		for _, cpu := range net.BelowThresholdCpuVars {
			if cpu < node.CpuVar {
				rank++
			}
		}

		// 检查除数是否为零
		lenBelow := len(net.BelowThresholdCpuVars)
		if lenBelow > 1 { // 防止除数为零
			CpuVarNormalization = -(1 - float64(rank-1)/float64(lenBelow-1))
		} else {
			CpuVarNormalization = 0 // 如果只有一个元素，归一化结果为0
		}
	}

	return CpuMeanNormalization, CpuVarNormalization
}

// 更新 CPU 均值虚拟队列
func (e *Evaluate) UpdateQMean() float64 {
	return math.Max(e.QMean+e.NormalCpuMean, 0)
}

// 更新 CPU 方差虚拟队列
func (e *Evaluate) UpdateQVar() float64 {
	return math.Max(e.QVar+e.NormalCpuVar, 0)
}

// 计算漂移加惩罚式子
func (e *Evaluate) DriftPlusPenalty() float64 {
	delayPart := e.Params.Weight * e.Delay
	meanPart := e.QMean * e.NormalCpuMean
	varPart := e.QVar * e.NormalCpuVar
	// 返回扩展公式总和
	return delayPart + meanPart + varPart
}
