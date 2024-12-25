package system_struct

import "time"

type ConfigInfo struct {
	PoolNum        int           //协程池数量
	DetectCycle    time.Duration //下发一次探测任务时长
	DetectNewNode  time.Duration // 检测节点周期
	CalculateCycle time.Duration // redis计算周期
	K              int           //路径数量
	Theta          float64       //惩罚系数
	Skip           int           //跳数限制
}
