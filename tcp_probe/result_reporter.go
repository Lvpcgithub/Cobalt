package tcp_probe

import (
	"bytes"
	"encoding/json" // 用于 JSON 编码和解码
	"fmt"           // 用于格式化字符串和输出
	"io"
	"log"
	"net/http" // 用于发送 HTTP 请求
)

// reportProbeResults 将单个探测结果上报到服务器
func reportProbeResults(url string, result ProbeResult) error {
	// 将单个探测结果编码为 JSON 格式的字节切片
	payload, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}

	log.Printf("Payload to send: %s", string(payload)) // 打印发送的 JSON 数据

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to report result: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // 读取服务器响应内容
		log.Printf("Server responded with: %s", string(body))
		return fmt.Errorf("server responded with status: %s", resp.Status)
	}

	return nil
}
