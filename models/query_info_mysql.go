package models

import (
	"Cobalt/system_struct"
	"database/sql"
	"fmt"
	"math"
)

// 查询设备名称
func QueryDeviceName(db *sql.DB) (*sql.Rows, error) {
	rows, err := db.Query("SELECT DISTINCT device_name FROM device_info")
	if err != nil {
		fmt.Println("Error executing query:", err)
		return nil, err
	}
	return rows, err
}
func GetIpByDevice(db *sql.DB) (map[string]string, error) {
	query := `
		SELECT device_name, MIN(ip_address) AS ip_address
		FROM ip_list
		GROUP BY device_name;
	`

	// 执行查询
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 存储结果
	result := make(map[string]string)
	for rows.Next() {
		var deviceName, ipAddress string
		if err := rows.Scan(&deviceName, &ipAddress); err != nil {
			return nil, err
		}
		// 将设备名与 IP 地址映射
		result[deviceName] = ipAddress
	}

	// 检查是否存在查询错误
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// 获取最新的 10 条 CPU 使用率数据，计算并返回均值和方差
func GetLatestCPUUsage(db *sql.DB, deviceName string) (*system_struct.CPUStats, error) {
	// SQL 查询语句，查询设备最新的 10 条数据的 CPU 使用率
	query := `
		SELECT cpu_usage
		FROM device_info
		WHERE device_name = ?
		ORDER BY created_at DESC
		LIMIT 10;
	`

	// 执行查询
	rows, err := db.Query(query, deviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	// 存储查询结果
	var cpuUsages []float64

	// 遍历结果集并提取 CPU 使用率
	for rows.Next() {
		var cpuUsage float64
		if err := rows.Scan(&cpuUsage); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		cpuUsages = append(cpuUsages, cpuUsage)
	}

	// 检查查询是否遇到错误
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}

	// 计算 CPU 使用率的均值和方差
	if len(cpuUsages) == 0 {
		return nil, fmt.Errorf("no CPU usage data found for device %s", deviceName)
	}

	// 计算均值 (mean)
	var sum float64
	for _, usage := range cpuUsages {
		sum += usage
	}
	mean := sum / float64(len(cpuUsages))

	// 计算方差 (variance)
	var varianceSum float64
	for _, usage := range cpuUsages {
		varianceSum += math.Pow(usage-mean, 2)
	}
	variance := varianceSum / float64(len(cpuUsages))

	// 创建一个结构体返回均值和方差
	stats := &system_struct.CPUStats{
		DeviceName: deviceName,
		Mean:       mean,
		Variance:   variance,
	}
	return stats, nil
}

// 根据 IP 地址查询设备名称
func GetDeviceNameByIP(db *sql.DB, ipAddress string) (string, error) {
	// SQL 查询语句，根据 IP 地址查找设备名称
	query := `
		SELECT device_name
		FROM ip_list
		WHERE ip_address = ?;
	`

	// 执行查询
	var deviceName string
	err := db.QueryRow(query, ipAddress).Scan(&deviceName)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有找到记录，返回空字符串和 nil 错误
			return "", nil
		}
		// 其他错误，返回错误信息
		return "", fmt.Errorf("failed to execute query: %v", err)
	}

	// 返回查询到的设备名称
	return deviceName, nil
}
