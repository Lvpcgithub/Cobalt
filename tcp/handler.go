package tcp

import (
	"log"
	"net"
)

// handleConnection 处理TCP连接
// 这个函数用于处理每个客户端的连接请求
// 它接受一个 net.Conn 类型的参数 conn，代表客户端的连接
func handleConnection(conn net.Conn) {
	// defer 语句确保在函数结束时自动关闭连接，释放资源
	defer conn.Close()

	// 使用 log.Printf 记录客户端的连接信息
	// conn.RemoteAddr().String() 获取连接客户端的远程地址（IP地址和端口号）
	log.Printf("Client connected: %s", conn.RemoteAddr().String())

	// 此处不对客户端发送的数据进行读取，也不返回任何数据给客户端
	// 函数执行结束后，连接会自动关闭
}
