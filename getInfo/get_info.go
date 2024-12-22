package getInfo

import (
	"Cobalt/dao"
	"Cobalt/models"
	"Cobalt/system_struct"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// 定义变量
var (
	detectData map[string]interface{}
	probeInfo  system_struct.ProbeResult
	rawData    map[string]interface{}
	deviceInfo system_struct.DeviceInfo
	ipInfo     system_struct.IpList
)

// 获取节点信息，并存储
func GetNodeInfo() *gin.Engine {
	r := gin.Default()
	// 定义变量
	r.POST("/fetch_node_info", func(c *gin.Context) {
		// 解析原始的 JSON 数据
		if err := c.ShouldBindJSON(&rawData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		// 将 map 转换为 JSON 字符串
		jsonData, err := json.Marshal(rawData)
		if err != nil {
			log.Fatalf("Failed to marshal raw data: %v", err)
		}
		// 将 JSON 字符串解析为结构体
		err = json.Unmarshal(jsonData, &deviceInfo)
		if err != nil {
			log.Fatalf("Failed to unmarshal JSON to struct: %v", err)
		}
		// 打印解析后的结构体数据
		fmt.Printf("解析后的结构体数据: %+v\n", deviceInfo)
		// 连接数据库
		db := dao.ConnectToDB()
		defer db.Close()
		// 插入信息到数据库
		if err := models.InsertDeviceInfo(db, deviceInfo); err != nil {
			log.Fatalf("Failed to insert system info: %v", err)
		}
		fmt.Println("System info inserted successfully!")
	})
	return r
}

// 获取节点IP信息
func GetNodeIpInfo() *gin.Engine {
	r := gin.Default()
	// 定义变量
	r.POST("/fetch_ip", func(c *gin.Context) {
		// 解析原始的 JSON 数据
		if err := c.ShouldBindJSON(&rawData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		// 将 map 转换为 JSON 字符串
		jsonData, err := json.Marshal(rawData)
		if err != nil {
			log.Fatalf("Failed to marshal raw data: %v", err)
		}
		// 将 JSON 字符串解析为结构体
		err = json.Unmarshal(jsonData, &ipInfo)
		if err != nil {
			log.Fatalf("Failed to unmarshal JSON to struct: %v", err)
		}
		// 打印解析后的结构体数据
		fmt.Printf("解析后的结构体数据: %+v\n", ipInfo)
		// 连接数据库
		db := dao.ConnectToDB()
		defer db.Close()
		// 插入信息到数据库
		if err := models.InsertIpInfo(db, ipInfo); err != nil {
			log.Fatalf("Failed to insert system info: %v", err)
		}
		fmt.Println("System info inserted successfully!")
	})
	return r
}

// 获取探测信息并存储
func GetDetectInfo() *gin.Engine {
	r := gin.Default()
	r.POST("/fetch_detect", func(c *gin.Context) {
		if err := c.ShouldBindJSON(&detectData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//转化数据
		jsonData, err := json.Marshal(detectData)
		if err != nil {
			log.Fatalf("Failed to marshal raw data: %v", err)
		}
		// 将 JSON 字符串解析为结构体
		err = json.Unmarshal(jsonData, &probeInfo)
		if err != nil {
			log.Fatalf("Failed to unmarshal JSON to struct: %v", err)
		}
		// 打印解析后的结构体数据
		fmt.Printf("解析后的结构体数据: %+v\n", probeInfo)
		//连接redis进行存储
		conn := dao.ConnRedis()
		defer conn.Close()
		models.CollectAndStoreData(conn, probeInfo)
	})
	return r
}
