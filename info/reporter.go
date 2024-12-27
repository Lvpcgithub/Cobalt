package info

import (
	"bytes"         // 用于处理字节缓冲区，方便数据处理
	"encoding/json" // 用于将结构体编码为JSON格式数据
	"fmt"           // 用于格式化字符串和输出
	"log"           // 用于日志记录，便于错误调试
	"net/http"      // 用于发送HTTP请求
	"time"          // 用于处理时间和定时任务
)

// APIURL 定义了用于接收系统信息的远程API地址
const APIURL = "http://124.70.34.63:8080/fetch_and_save"

// ReportSystemInfo 上报系统信息
// 接受一个 InfoData 类型的参数，将其转换为 JSON 格式并通过 HTTP POST 请求发送到指定的 APIURL
func ReportSystemInfo(info InfoData) {
	// 将系统信息编码为 JSON 格式
	jsonData, err := json.Marshal(info)
	if err != nil {
		// 如果编码失败，记录错误并终止程序
		log.Fatalf("Failed to marshal json data: %v", err)
	}

	// 发送HTTP POST请求，将 JSON 数据发送到 API
	resp, err := http.Post(APIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// 如果请求失败，记录错误并终止程序
		log.Fatalf("Failed to send data to API: %v", err)
	}

	// 确保响应体在函数结束时关闭，避免资源泄漏
	defer resp.Body.Close()

	// 输出服务器返回的响应状态
	fmt.Printf("Sent data to API, response status: %s\n", resp.Status)
}

// StartInfoCollector 启动信息收集和上报器
// 该函数会持续地收集系统信息，并每隔30秒将其上报到指定的API
func StartInfoCollector() {
	for {
		// 收集当前的系统信息
		info := CollectSystemInfo()

		// 将收集到的信息上报到API
		ReportSystemInfo(info)

		// 30秒，然后再次收集并上报系统信息
		time.Sleep(30 * time.Second)
	}
}
