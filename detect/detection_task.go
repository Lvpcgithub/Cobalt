package detect

import (
	"Cobalt/config"
	"Cobalt/dao"
	"Cobalt/models"
	"Cobalt/pool"
	"Cobalt/system_struct"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	ipAddresses []string
	deviceName  string
	ips         string
)

// 生成探测任务
func GenerateTasks() {
	//调用配置文件引出参数
	c := config.UseToml()
	// 初始化协程池
	pool.InitPool(c.PoolNum, func(task interface{}) {
		t := task.(system_struct.Task)
		printTask(t)
	})
	defer pool.ReleasePool() // 确保程序退出时释放协程池
	conn := dao.ConnRedis()
	defer conn.Close()
	//连接数据库并查询IP
	db := dao.ConnectToDB()
	defer db.Close()
	rows, err := models.QueryDeviceName(db)
	if err != nil {
		fmt.Println("Error rows", err)
	}
	//根据名称查询ip生成探测任务
	for rows.Next() {
		if err := rows.Scan(&deviceName); err != nil {
			fmt.Println("Error scanning row:", err)
			return
		}
		res, err := models.GetIpByDevice(db)
		if err != nil {
			fmt.Println("Error getting ip by device:", err)
			return
		}
		ipAddresses = append(ipAddresses, res[deviceName])
	}
	fmt.Printf("Unique IP Addresses: %v\n", ipAddresses)
	for {
		if len(ipAddresses) < 2 {
			log.Println("Not enough IP addresses to generate probing tasks.")
			return
		}
		// 立即发送一次探测任务
		sendProbingTasks(ipAddresses)
		// 定义每 30 秒下发一次探测任务的定时器/60秒检测是否有新的节点加入/每60s去redis取数据计算存到mysql
		ticker := time.NewTicker(c.DetectCycle * time.Second)
		tickerQuery := time.NewTicker(c.DetectNewNode * time.Second)
		tickerComputer := time.NewTicker(c.CalculateCycle * time.Second)
		defer ticker.Stop()
		defer tickerQuery.Stop()
		defer tickerComputer.Stop()
		// 使用通道监听定时器的触发事件
		for {
			select {
			case <-ticker.C:
				//case1：周期下发探测任务
				sendProbingTasks(ipAddresses)
				fmt.Println("探测任务下发：---》", ipAddresses)
			case <-tickerQuery.C:
				//****
				//case2：定时查询数据库，是否更新ip
				var newIpAddresses []string
				newRows, err := models.QueryDeviceName(db)
				if err != nil {
					fmt.Println("Error rows", err)
				}
				for newRows.Next() {
					if err := newRows.Scan(&ips); err != nil {
						fmt.Println("Error scanning row:", err)
						return
					}
					res, err := models.GetIpByDevice(db)
					if err != nil {
						newIpAddresses = append(newIpAddresses, res[ips])
					}
				}
				fmt.Println("定时器查询：", newIpAddresses)
				if len(newIpAddresses) != len(ipAddresses) {
					ipAddresses = append([]string(nil), newIpAddresses...) // 不一样则把新的切片赋值到ipaddress
				}
				fmt.Println("copy：", ipAddresses, newIpAddresses)
			case <-tickerComputer.C:
				//定时拿到数据并计算存到mysql里面去
				for i := 0; i < len(ipAddresses); i++ {
					for j := 0; j < len(ipAddresses); j++ {
						if i != j {
							models.RetrieveAndProcessData(conn, db, ipAddresses[i], ipAddresses[j])
						}
					}
				}
			}
		}
	}
}

// sendProbingTasks 生成并发送探测任务
func sendProbingTasks(ipAddresses []string) {
	mypool := pool.GetPool() // 获取协程池实例
	for i := 0; i < len(ipAddresses); i++ {
		for j := 0; j < len(ipAddresses); j++ {
			if i != j {
				task := system_struct.Task{
					SourceIP:      ipAddresses[i],
					DestinationIP: ipAddresses[j],
				}
				// 提交任务到协程池
				_ = mypool.Invoke(task)
			}
		}
	}
}

// sendTask 将任务发送到源 IP 所在机器的接口并打印探测信息
func sendTask(task system_struct.Task) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	// 目标机器的 API 地址
	apiURL := fmt.Sprintf("http://%s:8080/probe", task.SourceIP)
	taskData, err := json.Marshal(task)
	if err != nil {
		log.Printf("Failed to marshal task: %v", err)
		return
	}

	resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(taskData))
	if err != nil {
		log.Printf("Failed to send task to %s: %v", task.SourceIP, err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response from %s: %v", task.SourceIP, err)
		return
	}

	log.Printf("Response from %s -> %s: %d %s", task.SourceIP, task.DestinationIP, resp.StatusCode, string(body))

	if resp.StatusCode == http.StatusOK {
		log.Printf("Task successfully sent to %s", task.SourceIP)
	} else {
		log.Printf("Failed to send task to %s, status code: %d", task.SourceIP, resp.StatusCode)
	}
}

// 测试方法
func printTask(task system_struct.Task) {
	// 模拟随机超时
	taskData, err := json.Marshal(task)
	if err != nil {
		log.Printf("Failed to marshal task: %v", err)
		return
	}
	// 打印任务的 JSON 表示
	fmt.Printf("Generated Task: %s\n", taskData)
}
