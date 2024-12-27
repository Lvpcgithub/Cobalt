package tcp_probe

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	// 全局变量，用于控制探测任务的取消
	currentTaskCancel context.CancelFunc
	// 结果文件路径
)

func ProbeHandler(c *gin.Context) {
	var task ProbeTask

	// 绑定 JSON 数据到 task 结构体
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果未指定端口，使用默认端口 "50000"
	if task.Port == "" {
		task.Port = "50000"
	}

	// 取消当前正在执行的探测任务（如果存在）
	if currentTaskCancel != nil {
		currentTaskCancel()
	}

	// 清空探测结果文件
	err := clearProbeResultsFile(probeDataFilePath)
	if err != nil {
		log.Printf("Failed to clear probe results file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear previous results"})
		return
	}

	// 创建一个新的 context 和 cancel function
	ctx, cancel := context.WithCancel(context.Background())
	currentTaskCancel = cancel

	// 启动一个新的 Goroutine 来处理探测任务
	go func(task ProbeTask, ctx context.Context) {
		ticker := time.NewTicker(1 * time.Second)
		reportTicker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		defer reportTicker.Stop()

		var results []ProbeResult

		for {
			select {
			case <-ctx.Done():
				log.Println("Probe task cancelled")
				return

			case <-ticker.C:
				tcpDelay, err := performTCPProbe(task.IP2, task.Port)
				if err != nil {
					log.Printf("TCP probe failed: %v", err)
					continue
				}

				result := ProbeResult{
					IP1:       task.IP1,
					IP2:       task.IP2,
					TCPDelay:  tcpDelay / time.Millisecond, // 将延迟转换为毫秒
					Timestamp: time.Now(),
				}
				results = append(results, result)

			case <-reportTicker.C:
				if len(results) > 0 {
					// 计算所有探测结果的延迟平均值
					var totalDelay time.Duration
					for _, result := range results {
						totalDelay += result.TCPDelay
					}
					avgDelay := totalDelay / time.Duration(len(results))

					// 使用第一个结果的 IP1 和 IP2，创建一个新的探测结果用于上报
					avgResult := ProbeResult{
						IP1:       results[0].IP1,
						IP2:       results[0].IP2,
						TCPDelay:  avgDelay, // 已经是毫秒单位
						Timestamp: time.Now(),
					}

					// 将包含平均延迟的探测结果作为对象上报
					err := reportProbeResults("http://124.70.34.63:8080/fetch_detect", avgResult)
					if err != nil {
						log.Printf("Failed to report probe results: %v", err)
					} else {
						// 如果上报成功，清空探测结果文件
						err := clearProbeResultsFile(probeDataFilePath)
						if err != nil {
							log.Printf("Failed to clear probe results file after reporting: %v", err)
						}
					}
					// 清空内存中的结果切片
					results = []ProbeResult{}
				}
			}
		}
	}(task, ctx)

	c.JSON(http.StatusOK, gin.H{"status": "probe started"})
}

// clearProbeResultsFile 清空探测结果文件内容
func clearProbeResultsFile(filePath string) error {
	// 打开文件并清空内容
	file, err := os.OpenFile(filePath, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

// performTCPProbe 执行TCP探测
// 这个函数尝试连接指定的 IP 和端口，并测量连接的延迟时间
func performTCPProbe(ip, port string) (time.Duration, error) {
	// 记录开始时间
	start := time.Now()
	// 设置连接超时时间为 5 秒
	timeout := 5 * time.Second
	// 尝试在指定的 IP 和端口上建立 TCP 连接
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), timeout)
	if err != nil {
		// 如果连接失败，返回错误信息
		return 0, err
	}
	// 确保连接在函数结束时关闭，释放资源
	defer conn.Close()

	// 返回连接时间的持续时间和 nil 错误
	return time.Since(start), nil
}
