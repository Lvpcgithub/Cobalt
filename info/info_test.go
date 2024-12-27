package info

import (
	"encoding/json"
	"testing"
)

// TestCollectSystemInfo 测试 CollectSystemInfo 函数收集的信息
func TestCollectSystemInfo(t *testing.T) {
	// 调用 CollectSystemInfo 函数来收集信息
	info := CollectSystemInfo()

	// 将收集到的信息编码为 JSON 格式，以便查看其具体内容
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal system info to JSON: %v", err)
	}

	// 打印收集到的系统信息的 JSON 表示
	t.Logf("Collected System Info: %s", jsonData)

	// 这里可以进行更多的断言检查收集到的信息是否符合预期
	if info.IP == "" {
		t.Errorf("Expected non-empty IP, but got an empty string")
	}

	// 根据具体的字段进行更多检查
	// 例如：检查 CPUInfo, MemoryInfo 等字段的合理性
}
