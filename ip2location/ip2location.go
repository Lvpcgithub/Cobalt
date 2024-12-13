package ip2location

import (
	"Cobalt/dao"
	"Cobalt/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ip2location/ip2location-go/v9"
	"net/http"
)

func ip2Location() {
	// 初始化数据库连接
	db := dao.ConnectToDB()
	defer db.Close()

	// 初始化IP2Location数据库
	ipDb, err := ip2location.OpenDB("../IP2LOCATION-LITE-DB11.BIN")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ipDb.Close()
	// 初始化Gin路由
	r := gin.Default()
	r.LoadHTMLGlob("../templates/*")

	r.GET("/", func(c *gin.Context) {
		// 查询最新的IP地址
		ip, err := models.GetLatestIP(db)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{"Error": "Failed to get IP"})
			return
		}

		// 使用IP2Location模块查询IP信息
		fmt.Println(ip)
		results, err := ipDb.Get_all(ip)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{"Error": "Failed to query IP location"})
			return
		}

		// 渲染结果
		c.HTML(http.StatusOK, "index.html", gin.H{"Result": results})
	})

	r.POST("/lookup", func(c *gin.Context) {
		ip := c.PostForm("ip")
		results, err := ipDb.Get_all(ip)
		if err != nil {
			fmt.Println(err)
			return
		}
		c.HTML(http.StatusOK, "index.html", gin.H{"Result": results})
	})

	r.Run(":8081")
}
