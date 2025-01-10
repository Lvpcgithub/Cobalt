package models

import (
	"Cobalt/system_struct"
	"database/sql"
	"fmt"
	"math"
	"sort"
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

func GetCpuAvgAndVariance(db *sql.DB, thresholdCpuMean float64, thresholdCpuVar float64) (aboveCpuMeans, belowCpuMeans, aboveCpuVars, belowCpuVars []float64, err error) {
	// 查询数据库，获取每个设备的最近 10 条记录的 CPU 使用率均值和方差
	query := `
	WITH RankedDevices AS (
		SELECT 
			device_name, 
			cpu_usage,
			timestamp,
			ROW_NUMBER() OVER (PARTITION BY device_name ORDER BY timestamp DESC) AS rn
		FROM device_info
	)
	SELECT 
		device_name,
		AVG(cpu_usage) AS avg_cpu_usage,
		VARIANCE(cpu_usage) AS variance_cpu_usage
	FROM RankedDevices
	WHERE rn <= 10
	GROUP BY device_name;
	`

	// 执行查询
	rows, queryErr := db.Query(query)
	if queryErr != nil {
		return nil, nil, nil, nil, queryErr
	}
	defer rows.Close()

	// 遍历查询结果
	for rows.Next() {
		var deviceName string
		var avgCpuUsage float64
		var varianceCpuUsage float64

		// 将每行查询结果扫描到变量
		scanErr := rows.Scan(&deviceName, &avgCpuUsage, &varianceCpuUsage)
		if scanErr != nil {
			return nil, nil, nil, nil, scanErr
		}

		// 分类到相应的数组
		if avgCpuUsage > thresholdCpuMean {
			aboveCpuMeans = append(aboveCpuMeans, avgCpuUsage)
		} else {
			belowCpuMeans = append(belowCpuMeans, avgCpuUsage)
		}

		if varianceCpuUsage > thresholdCpuVar {
			aboveCpuVars = append(aboveCpuVars, varianceCpuUsage)
		} else {
			belowCpuVars = append(belowCpuVars, varianceCpuUsage)
		}
	}

	// 检查查询是否有错误
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, nil, nil, nil, rowsErr
	}

	// 对每个数组进行升序排序
	sort.Float64s(aboveCpuMeans)
	sort.Float64s(belowCpuMeans)
	sort.Float64s(aboveCpuVars)
	sort.Float64s(belowCpuVars)

	return aboveCpuMeans, belowCpuMeans, aboveCpuVars, belowCpuVars, nil
}

func QueryDeviceIPs(db *sql.DB) ([]system_struct.DeviceUseInfo, error) {
	// SQL 查询语句
	query := `
		SELECT 
			ip_list.device_name,
			GROUP_CONCAT(DISTINCT ip_list.ip_address ORDER BY ip_list.ip_address ASC SEPARATOR ',') AS ips,
			di.cpu_cores,
			di.cpu_model_name,
			di.cpu_mhz,
			di.memory_total,
			di.memory_used,
			di.memory_used_percent
		FROM 
			ip_list
		JOIN 
			device_info di
		ON 
			ip_list.device_name = di.device_name
		WHERE 
			di.created_at = (
				SELECT 
					MAX(created_at)
				FROM 
					device_info
				WHERE 
					device_info.device_name = di.device_name
			)
		GROUP BY 
    		ip_list.device_name, di.cpu_cores, di.cpu_model_name, di.cpu_mhz, di.memory_total, di.memory_used, di.memory_used_percent;
			`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var devices []system_struct.DeviceUseInfo
	for rows.Next() {
		var device system_struct.DeviceUseInfo
		err := rows.Scan(&device.DeviceName, &device.IPs, &device.CPUCores, &device.CPUModelName, &device.CPUMHz, &device.MemoryTotal, &device.MemoryUsed, &device.MemoryUsedPercent)
		if err != nil {
			return nil, fmt.Errorf("row scan failed: %v", err)
		}
		devices = append(devices, device)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	return devices, nil
}
