package handler

import (
	"bufio"
	"demo1/proxy/config"
	"demo1/proxy/connection"
	smux2 "demo1/proxy/smux_usage"
	"github.com/gin-gonic/gin"
	"github.com/xtaci/smux"
	"log"
	"net"
	"net/http"
	"sync"
)

var (
	// 用来存储不同IP的连接池
	connectionPools = make(map[string]*connection.Pool)
	// 用来缓存当前使用的TCP连接和SMUX会话
	cachedConn    net.Conn
	cachedSession *smux.Session
	// 用来保护连接池 map 的并发访问
	mu     sync.Mutex
	stream *smux.Stream
)

func HTTPRequestHandler(c *gin.Context, tcpPool *connection.Pool) {

	// 创建一个示例 Packet 对象
	originalPacket := &config.Packet{
		Length:      64,
		HeaderLen:   20,
		Timestamp:   1672531200,
		PacketID:    12345678,
		PacketType:  1,
		Property:    256,
		Priority:    5,
		HopCounts:   2,
		PacketCount: 1,
		Offsets:     []uint8{10},
		Padding:     []uint8{0, 0, 0, 0},
		HopList:     []uint32{3232235777, 3232235778}, // 对应 IP: 192.168.1.1 和 192.168.1.2
	}

	header, err := config.SerializePacket(originalPacket)

	// 判断下一跳是否为服务器
	if IsServer(nextHop) {
		// 如果下一跳是服务器，则直接建立一个 TCP 连接
		conn, err := net.Dial("tcp", nextHop+":8080")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to establish connection to server",
			})
			return
		}
		defer conn.Close()

		// 直接使用 TCP 连接转发 HTTP 请求
		err = ForwardRequestWithoutSMUX(conn, c.Request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to forward request to server",
			})
			return
		}
	} else {
		// 获取连接池
		tcpPool, err = GetOrCreateConnectionPool(nextHop)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get or create connection pool",
			})
			return
		}

		// 从连接池获取一个连接
		conn, err := tcpPool.Get()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get connection from pool",
			})
			return
		}

		// **从连接池中获取或创建SMUX会话**
		session, err := smux2.GetOrCreateSMUXSession(conn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create or get SMUX session",
			})
			return
		}

		err = ForwardRequestWithSMUX(session, c.Request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to forward request to server",
			})
			return
		}
	}
}

// IsServer 判断下一跳是否为服务器
func IsServer(nextHopIP string) bool {
	// 这里你可以定义你的逻辑来确定是否是服务器
	return nextHopIP == config.routeTable.ServerIP
}

// 创建或获取到下一跳的连接池
func GetOrCreateConnectionPool(nextHopIP string) (*connection.Pool, error) {
	mu.Lock()
	defer mu.Unlock()

	// 检查是否已经有针对下一跳的连接池
	if pool, exists := connectionPools[nextHopIP]; exists {
		return pool, nil
	}
	factory := func() (net.Conn, error) { return net.Dial("tcp", nextHopIP) }
	// 如果连接池不存在，则为该 IP 创建新的连接池
	tcpPool, err := connection.NewChannelPool(5, 20, factory)
	if err != nil {
		log.Printf("Error creating connection pool for %s: %v", nextHopIP, err)
		return nil, err
	}

	// 将新创建的连接池保存到 map 中
	connectionPools[nextHopIP] = tcpPool
	log.Printf("Created new connection pool for next hop %s", nextHopIP)
	return tcpPool, nil
}

// SetupRouter 配置 Gin 路由
func SetupRouter(tcpPool *connection.Pool) *gin.Engine {
	// 初始化 Gin 引擎
	router := gin.Default()

	// 定义 GET 和 POST 路由
	router.GET("/", func(c *gin.Context) {
		HTTPRequestHandler(c, tcpPool)
	})

	router.POST("/", func(c *gin.Context) {
		HTTPRequestHandler(c, tcpPool)
	})

	return router
}

// ForwardRequestWithSMUX 使用 SMUX 流转发 HTTP 请求
func ForwardRequestWithSMUX(session *smux.Session, req *http.Request) error {
	// 打开一个新的 SMUX 流
	stream, err := smux2.OpenSMUXStream(session)
	if err != nil {
		log.Printf("Failed to open SMUX stream: %v", err)
		return err
	}
	defer stream.Close()

	// 将 HTTP 请求写入 SMUX 流
	err = req.Write(stream)
	if err != nil {
		log.Printf("Failed to write request to SMUX stream: %v", err)
		return err
	}

	log.Println("HTTP request written to SMUX stream")

	// 从 SMUX 流中读取响应
	resp, err := http.ReadResponse(bufio.NewReader(stream), req)
	if err != nil {
		log.Printf("Failed to read response from SMUX stream: %v", err)
		return err
	}

	// 打印响应状态
	log.Printf("Received response with status: %s", resp.Status)

	// 处理响应体
	defer resp.Body.Close()
	return nil
}

func ForwardRequestWithoutSMUX(conn net.Conn, req *http.Request) error {
	// 将 HTTP 请求写入 TCP 连接
	err := req.Write(conn)
	if err != nil {
		log.Printf("Failed to write request to TCP connection: %v", err)
		return err
	}

	log.Println("HTTP request written to TCP connection")

	// 从 TCP 连接中读取响应
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		log.Printf("Failed to read response from TCP connection: %v", err)
		return err
	}

	// 打印响应状态
	log.Printf("Received response with status: %s", resp.Status)

	// 处理响应体并关闭它
	defer resp.Body.Close()

	return nil
}
