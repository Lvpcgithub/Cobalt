package tcp_probe

import (
	"encoding/json" // 用于 JSON 编码和解码
	"fmt"           // 用于格式化字符串和输出
	"os"            // 提供对操作系统功能的访问，如文件操作
	"sync"          // 提供同步原语，如互斥锁（Mutex）
)

var (
	probeDataFilePath = "probe_results.json" // 定义探测结果保存的文件路径
	probeMutex        sync.Mutex             // 定义互斥锁，用于保护文件写入的并发安全
)

// saveProbeResult 保存探测结果到本地文件
// 参数 results 是一个可变参数列表，表示多个探测结果
func saveProbeResult(results ...ProbeResult) {
	probeMutex.Lock()         // 加锁，确保在多线程环境下对文件的写操作是线程安全的
	defer probeMutex.Unlock() // 函数返回前解锁

	// 打开或创建探测数据文件，以追加模式写入文件
	// os.O_CREATE：如果文件不存在则创建
	// os.O_WRONLY：以只写模式打开文件
	// os.O_APPEND：以追加模式打开文件
	// 0644：文件权限，表示文件所有者可读写，组用户和其他用户只读
	file, err := os.OpenFile(probeDataFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// 如果打开文件失败，输出错误信息并返回
		fmt.Printf("Failed to open probe data file: %v\n", err)
		return
	}
	defer file.Close() // 函数返回前关闭文件，释放资源

	// 遍历所有的探测结果，并将其编码为 JSON 格式写入文件
	for _, result := range results {
		if err := json.NewEncoder(file).Encode(result); err != nil {
			// 如果写入文件失败，输出错误信息
			fmt.Printf("Failed to write probe data to file: %v\n", err)
		}
	}
}
