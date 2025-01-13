package deviceApi

import (
	"Cobalt/dao"
	"Cobalt/models"
	"Cobalt/system_struct"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	deviceData  []system_struct.DeviceUseInfo
	dataLock    sync.RWMutex // 用于保证并发安全
	updateError error        // 保存最近一次更新的错误信息
)

func DeviceInfo() *gin.Engine {
	// 创建 Gin 路由
	r := gin.Default()
	db := dao.ConnectToDB()
	defer db.Close()
	// **程序启动时立即查询一次设备信息**
	queryAndUpdateDeviceData(db)

	// 启动协程池的定时更新任务
	go updateDeviceData(db)
	// 定义设备信息接口
	r.GET("/devices", func(c *gin.Context) {
		dataLock.RLock()
		defer dataLock.RUnlock()

		// 如果最近更新时发生错误，返回错误信息
		if updateError != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to update device information: %v", updateError),
			})
			return
		}

		// 返回最新的设备信息
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"devices": deviceData,
		})
	})

	return r
}

// 查询并更新全局设备信息的方法
func queryAndUpdateDeviceData(db *sql.DB) {

	// 查询数据库
	devices, err := models.QueryDeviceIPs(db)
	if err != nil {
		log.Printf("failed to query device information: %v", err)
		// 保存错误信息供接口显示
		dataLock.Lock()
		updateError = err
		dataLock.Unlock()
		return
	}

	// 更新全局设备信息
	dataLock.Lock()
	deviceData = devices
	updateError = nil // 清除之前的错误
	dataLock.Unlock()

}

// 定时更新设备信息的方法
func updateDeviceData(db *sql.DB) {
	ticker := time.NewTicker(1 * time.Minute) // 每 1 分钟执行一次
	defer ticker.Stop()

	for range ticker.C {
		queryAndUpdateDeviceData(db) // 提交任务到协程池
	}
}

// 开放topology信息
func TopologyInfo() *gin.Engine {
	r := gin.Default()

	r.GET("/topology", func(c *gin.Context) {
		models.Mu.Lock()
		defer models.Mu.Unlock()

		// 返回 topology 数据
		c.JSON(200, models.Topology)
	})

	return r
}
