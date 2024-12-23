package test_info

import (
	"Cobalt/dao"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type LinkInfo struct {
	IP1        string `json:"ip1"`
	IP2        string `json:"ip2"`
	LinkWeight int    `json:"link_weight"`
}

func CaculateInfo() *gin.Engine {
	db := dao.ConnectToDB()
	defer db.Close()
	r := gin.Default()
	// Cache data in memory
	var linkInfos []LinkInfo
	rows, err := db.Query("SELECT ip1, ip2, link_weight FROM global_link_info")
	if err != nil {
		log.Fatal("Database query error:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var linkInfo LinkInfo
		if err := rows.Scan(&linkInfo.IP1, &linkInfo.IP2, &linkInfo.LinkWeight); err != nil {
			log.Fatal("Data scan error:", err)
		}
		linkInfos = append(linkInfos, linkInfo)
	}

	// Define the endpoint
	r.GET("/topology_info", func(c *gin.Context) {
		c.JSON(http.StatusOK, linkInfos)
	})
	return r
}
func CaculateInfo_1() *gin.Engine {
	db := dao.ConnectToDB()
	defer db.Close()
	r := gin.Default()
	// Cache data in memory
	var linkInfos []LinkInfo
	rows, err := db.Query("SELECT ip1, ip2, link_weight FROM global_link_info_1")
	if err != nil {
		log.Fatal("Database query error:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var linkInfo LinkInfo
		if err := rows.Scan(&linkInfo.IP1, &linkInfo.IP2, &linkInfo.LinkWeight); err != nil {
			log.Fatal("Data scan error:", err)
		}
		linkInfos = append(linkInfos, linkInfo)
	}

	// Define the endpoint
	r.GET("/topology_info_1", func(c *gin.Context) {
		c.JSON(http.StatusOK, linkInfos)
	})
	return r
}
func CaculateInfo_2() *gin.Engine {
	db := dao.ConnectToDB()
	defer db.Close()
	r := gin.Default()
	// Cache data in memory
	var linkInfos []LinkInfo
	rows, err := db.Query("SELECT ip1, ip2, link_weight FROM global_link_info_2")
	if err != nil {
		log.Fatal("Database query error:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var linkInfo LinkInfo
		if err := rows.Scan(&linkInfo.IP1, &linkInfo.IP2, &linkInfo.LinkWeight); err != nil {
			log.Fatal("Data scan error:", err)
		}
		linkInfos = append(linkInfos, linkInfo)
	}

	// Define the endpoint
	r.GET("/topology_info_2", func(c *gin.Context) {
		c.JSON(http.StatusOK, linkInfos)
	})
	return r
}
