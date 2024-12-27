package tcp

import (
	"log"
	"net"
)

// StartTCPServer 启动一个简单的TCP服务器
// 这个函数接受一个端口号作为参数，并在该端口上启动TCP服务器
func StartTCPServer(port string) {
	// 使用 net.Listen 在指定的端口上启动TCP监听器
	// 第一个参数 "tcp" 表示协议，第二个参数是监听的地址和端口
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		// 如果监听失败，记录错误并终止程序执行
		log.Fatalf("Failed to start TCP server on port %s: %v", port, err)
	}
	// defer 语句确保在函数返回时关闭监听器，释放资源
	defer listener.Close()

	// 记录日志，表示服务器已经在指定端口上开始监听
	log.Printf("TCP server is listening on port %s", port)

	// 无限循环，持续等待和接受新的连接
	for {
		// Accept 方法阻塞并等待客户端的连接请求
		conn, err := listener.Accept()
		if err != nil {
			// 如果接收连接失败，记录错误并继续等待下一个连接
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// 成功接收连接后，启动一个新的 goroutine 来处理该连接
		// 使用 goroutine 可以并发处理多个客户端连接
		go handleConnection(conn)
	}
}
