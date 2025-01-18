package models

import (
	"Cobalt/system_struct"
	"encoding/json"
	"fmt"
	"os"
)

// 保存 Topology 到文件
func SaveTopologyToFile() error {
	file, err := os.Create("topology.json") // 如果文件不存在，创建新文件
	if err != nil {
		return fmt.Errorf("failed to create topology file: %v", err)
	}
	defer file.Close()

	// 使用 JSON 格式写入 Topology 数据
	encoder := json.NewEncoder(file)
	err = encoder.Encode(Topology)
	if err != nil {
		return fmt.Errorf("failed to encode topology to file: %v", err)
	}

	return nil
}

// 从本地文件读取 Topology 数据
func LoadTopologyFromFile() (*system_struct.TopologyMatrix, error) {
	file, err := os.Open("topology.json") // 打开文件
	if err != nil {
		return nil, fmt.Errorf("failed to open topology file: %v", err)
	}
	defer file.Close()

	var topology system_struct.TopologyMatrix
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&topology) // 将文件内容解码到 Topology 结构体中
	if err != nil {
		return nil, fmt.Errorf("failed to decode topology from file: %v", err)
	}

	return &topology, nil
}
