package handler

import (
	"bytes"
	"fmt"
	"github.com/xtaci/smux" // 使用 SMUX 协议库
	"io"
	"net"
)

// Module2API: 模块2的对外接口
type Module2API struct {
	ClientServerAPI *Module1API // 模块1的接口实例
}

// NewModule2API: 创建模块2实例
func NewModule2API(clientServerAPI *Module1API) *Module2API {
	return &Module2API{ClientServerAPI: clientServerAPI}
}

// StartProxyServer: 启动代理节点服务器
func (api *Module2API) StartProxyServer(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start proxy server: %w", err)
	}
	defer listener.Close()

	fmt.Printf("ProxyNode: Listening on %s\n", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}

		// 处理代理节点连接
		go api.handleProxyConnection(conn)
	}
}

// handleProxyConnection: 处理代理节点连接
func (api *Module2API) handleProxyConnection(conn net.Conn) {
	defer conn.Close()

	// 初始化 SMUX 会话
	session, err := smux.Server(conn, nil)
	if err != nil {
		fmt.Println("Failed to create SMUX session:", err)
		return
	}
	defer session.Close()

	for {
		// 接收 SMUX 流
		stream, err := session.AcceptStream()
		if err != nil {
			fmt.Println("Failed to accept stream:", err)
			return
		}

		// 处理 SMUX 流数据
		go api.handleStream(stream)
	}
}

// handleStream: 处理 SMUX 流
func (api *Module2API) handleStream(stream *smux.Stream) {
	defer stream.Close()

	// 读取请求数据
	data, err := io.ReadAll(stream)
	if err != nil {
		fmt.Println("Failed to read stream data:", err)
		return
	}

	// 解封数据并解析为请求信息
	method, targetURL, headers, body := api.parseRequest(data)

	// 判断目标：如果是服务器，交给模块1处理；如果是代理节点，转发到下一跳
	if api.isServerTarget(targetURL) {
		// 转发到目标服务器
		responseBytes, err := api.ClientServerAPI.forwardToServer(data, targetURL)
		if err != nil {
			fmt.Println("Failed to forward to server:", err)
			stream.Write([]byte(fmt.Sprintf("Error: %v", err)))
			return
		}

		// 返回响应给请求方
		stream.Write(responseBytes)
	} else {
		// 转发到下一跳代理节点
		responseBytes, err := api.SendRequestToProxy(targetURL, method, targetURL, headers, body)
		if err != nil {
			fmt.Println("Failed to forward to next proxy:", err)
			stream.Write([]byte(fmt.Sprintf("Error: %v", err)))
			return
		}

		// 返回响应给请求方
		stream.Write(responseBytes)
	}
}

// parseRequest: 解析 SMUX 流中的请求数据
func (api *Module2API) parseRequest(data []byte) (method, targetURL string, headers map[string][]string, body []byte) {
	// 示例实现，解析请求格式：第一行为方法和URL，接下来为头部，最后为Body
	headers = make(map[string][]string)
	parts := bytes.Split(data, []byte("\n\n")) // 分为头部和Body
	headerPart := bytes.Split(parts[0], []byte("\n"))

	// 第一行是方法和URL
	firstLine := bytes.Split(headerPart[0], []byte(" "))
	method = string(firstLine[0])
	targetURL = string(firstLine[1])

	// 剩余部分是头部
	for _, line := range headerPart[1:] {
		header := bytes.SplitN(line, []byte(": "), 2)
		if len(header) == 2 {
			key := string(header[0])
			value := string(header[1])
			headers[key] = append(headers[key], value)
		}
	}

	// Body部分
	if len(parts) > 1 {
		body = parts[1]
	}

	return
}

// isServerTarget: 判断目标是否是服务器
func (api *Module2API) isServerTarget(targetURL string) bool {
	// 示例实现：根据目标URL判断是否为服务器
	// 如果是以 "http://" 或 "https://" 开头，则认为是服务器
	return bytes.HasPrefix([]byte(targetURL), []byte("http://")) || bytes.HasPrefix([]byte(targetURL), []byte("https://"))
}

// SendRequestToProxy: 将请求发送到代理节点
func (api *Module2API) SendRequestToProxy(proxyAddr, method, url string, headers map[string][]string, body []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %w", err)
	}
	defer conn.Close()

	// 创建 SMUX 会话
	session, err := smux.Client(conn, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMUX session: %w", err)
	}
	defer session.Close()

	// 创建 SMUX 流
	stream, err := session.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("failed to open SMUX stream: %w", err)
	}
	defer stream.Close()

	// 构造请求数据
	requestData := fmt.Sprintf("%s %s\n", method, url)
	for key, values := range headers {
		for _, value := range values {
			requestData += fmt.Sprintf("%s: %s\n", key, value)
		}
	}
	requestData += "\n" + string(body)

	// 发送数据
	_, err = stream.Write([]byte(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to write to stream: %w", err)
	}

	// 接收响应
	response, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	return response, nil
}
