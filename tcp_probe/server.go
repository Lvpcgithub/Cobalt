package tcp_probe

import "github.com/gin-gonic/gin" // 导入 Gin 框架，用于创建和管理 HTTP 服务器

// StartAPIServer 启动API服务器
// 这个函数使用 Gin 框架创建一个简单的 HTTP 服务器，并在指定的端口上运行
func StartAPIServer() {
	// 创建一个默认的 Gin 路由器
	// gin.Default() 返回一个默认的路由器实例，包含了 Logger 和 Recovery 中间件
	router := gin.Default()

	// 定义一个 POST 路由，用于处理 TCP 探测任务
	// 当客户端向 "/probe" 端点发送 POST 请求时，ProbeHandler 函数将处理该请求
	router.POST("/probe", ProbeHandler)

	// 启动服务器，监听指定的端口（8080）
	// router.Run(":8080") 会在本地的 8080 端口上启动 HTTP 服务器，并开始监听请求
	router.Run(":8080")
}
