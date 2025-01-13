package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

// PacketHeader 定义自定义包头
type PacketHeader struct {
	Length      uint16
	HeaderLen   uint16
	Timestamp   uint32
	PacketID    uint32
	PacketType  uint8
	Priority    uint8
	Property    uint16
	HopCounts   uint8
	PacketCount uint8
	Offsets     []uint8
	Padding     []byte
	HopList     []uint32
}

// serializeHeader 将包头序列化为字节流
func serializeHeader(header PacketHeader, data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	// 写入固定字段
	if err := binary.Write(buf, binary.BigEndian, header.Timestamp); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.PacketID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.PacketType); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.Priority); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.Property); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.HopCounts); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.PacketCount); err != nil {
		return nil, err
	}
	// 写入可变字段
	if err := binary.Write(buf, binary.BigEndian, header.Offsets); err != nil {
		return nil, err
	}
	// 计算 Padding 长度
	paddingLen := (4 - (buf.Len() % 4)) % 4
	header.Padding = make([]byte, paddingLen)
	if err := binary.Write(buf, binary.BigEndian, header.Padding); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.HopList); err != nil {
		return nil, err
	}
	// 计算 HeaderLen
	headerLen := uint16(buf.Len())
	// 创建最终的数据包
	finalBuf := new(bytes.Buffer)
	// 写入 Length
	length := uint16(2 + 2 + int(headerLen) + len(data))
	if err := binary.Write(finalBuf, binary.BigEndian, length); err != nil {
		return nil, err
	}
	// 写入 HeaderLen
	if err := binary.Write(finalBuf, binary.BigEndian, headerLen); err != nil {
		return nil, err
	}
	// 写入包头数据
	finalBuf.Write(buf.Bytes())
	// 写入数据部分
	finalBuf.Write(data)
	return finalBuf.Bytes(), nil
}

// deserializeHeader 将字节流反序列化为包头
func deserializeHeader(headerData []byte, headerLen uint16) (PacketHeader, error) {
	var header PacketHeader
	buf := bytes.NewReader(headerData)

	// 读取固定字段
	if err := binary.Read(buf, binary.BigEndian, &header.Timestamp); err != nil {
		return PacketHeader{}, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.PacketID); err != nil {
		return PacketHeader{}, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.PacketType); err != nil {
		return PacketHeader{}, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.Priority); err != nil {
		return PacketHeader{}, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.Property); err != nil {
		return PacketHeader{}, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.HopCounts); err != nil {
		return PacketHeader{}, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.PacketCount); err != nil {
		return PacketHeader{}, err
	}
	// 读取可变字段
	header.Offsets = make([]uint8, header.PacketCount)
	if err := binary.Read(buf, binary.BigEndian, &header.Offsets); err != nil {
		return PacketHeader{}, err
	}
	// 计算 Padding 长度
	paddingLen := (4 - (headerLen % 4)) % 4
	header.Padding = make([]byte, paddingLen)
	if err := binary.Read(buf, binary.BigEndian, &header.Padding); err != nil {
		return PacketHeader{}, err
	}
	// 读取 HopList
	header.HopList = make([]uint32, header.HopCounts)
	if err := binary.Read(buf, binary.BigEndian, &header.HopList); err != nil {
		return PacketHeader{}, err
	}
	return header, nil
}

// startProxyServer1 启动代理服务器1（HTTP 服务器）
func startProxyServer1() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 序列化整个 HTTP 请求
		var requestBuffer bytes.Buffer
		err := r.Write(&requestBuffer)
		if err != nil {
			log.Printf("序列化 HTTP 请求失败: %v", err)
			http.Error(w, "序列化请求失败", http.StatusInternalServerError)
			return
		}
		requestData := requestBuffer.Bytes()
		log.Printf("接收到 HTTP 请求: %s %s, 请求数据长度: %d bytes", r.Method, r.URL.Path, len(requestData))

		// 创建自定义包头
		header := PacketHeader{
			HeaderLen:   0, // 由 serializeHeader 计算
			Timestamp:   uint32(time.Now().Unix()),
			PacketID:    123456,
			PacketType:  1,
			Priority:    2,
			Property:    300,
			HopCounts:   3,
			PacketCount: 1,
			Offsets:     []uint8{10},
			// Padding 会在 serializeHeader 中计算并填充
			HopList: []uint32{
				binary.BigEndian.Uint32(net.ParseIP("192.168.0.1").To4()),
				binary.BigEndian.Uint32(net.ParseIP("192.168.0.2").To4()),
				binary.BigEndian.Uint32(net.ParseIP("192.168.0.3").To4()),
			},
		}

		// 序列化包头
		headerData, err := serializeHeader(header, requestData)
		if err != nil {
			log.Printf("序列化包头失败: %v", err)
			http.Error(w, "序列化包头失败", http.StatusInternalServerError)
			return
		}

		// 将包头和请求数据拼接
		fullData := headerData
		log.Printf("拼接后的完整数据长度: %d", len(fullData))

		// 连接到代理服务器2
		proxy2Conn, err := net.Dial("tcp", "127.0.0.1:8082")
		if err != nil {
			log.Printf("连接代理服务器2失败: %v", err)
			http.Error(w, "连接代理服务器2失败", http.StatusInternalServerError)
			return
		}
		defer proxy2Conn.Close()
		log.Println("成功连接到代理服务器2")

		// 发送数据到代理服务器2
		_, err = proxy2Conn.Write(fullData)
		if err != nil {
			log.Printf("发送数据到代理服务器2失败: %v", err)
			http.Error(w, "发送数据到代理服务器2失败", http.StatusInternalServerError)
			return
		}
		log.Println("数据已发送到代理服务器2")

		// 设置读取超时
		proxy2Conn.SetReadDeadline(time.Now().Add(10 * time.Second))

		// 读取代理服务器2的响应
		responseData, err := io.ReadAll(proxy2Conn)
		if err != nil {
			log.Printf("读取代理服务器2响应失败: %v", err)
			http.Error(w, "读取代理服务器2响应失败", http.StatusInternalServerError)
			return
		}
		log.Printf("接收到代理服务器2的响应数据: %s", string(responseData))

		// 将响应返回给客户端
		// 注意：responseData 已经包含了完整的HTTP响应，因此直接写入响应体即可
		// 不需要调用 w.WriteHeader，因为响应头已经包含在 responseData 中
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("Hello, i am server!"))
		if err != nil {
			log.Printf("响应已返回给客户端失败: %v", err)
			return
		}
		log.Println("响应已返回给客户端")
	})

	// 启动 HTTP 服务器
	log.Println("代理服务器1（HTTP 服务器）已启动，监听 127.0.0.1:8081")
	if err := http.ListenAndServe("127.0.0.1:8081", nil); err != nil {
		log.Fatalf("代理服务器1启动失败: %v", err)
	}
}

// startProxyServer2 启动代理服务器2
func startProxyServer2() {
	// 监听本地端口 8082
	listener, err := net.Listen("tcp", "127.0.0.1:8082")
	if err != nil {
		log.Fatalf("代理服务器2启动失败: %v", err)
	}
	defer listener.Close()
	log.Println("代理服务器2已启动，监听 127.0.0.1:8082")

	for {
		// 接受代理服务器1的连接
		proxy1Conn, err := listener.Accept()
		if err != nil {
			log.Printf("接受代理服务器1连接失败: %v", err)
			continue
		}
		log.Printf("接收到来自 %s 的新连接", proxy1Conn.RemoteAddr().String())
		go handleProxy1Request(proxy1Conn)
	}
}

// handleProxy1Request 处理代理服务器1的请求
func handleProxy1Request(proxy1Conn net.Conn) {
	defer proxy1Conn.Close()

	buf := bufio.NewReader(proxy1Conn)

	// 读取 Length
	var length uint16
	err := binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		log.Printf("读取 Length 失败: %v", err)
		return
	}
	// 读取 HeaderLen
	var headerLen uint16
	err = binary.Read(buf, binary.BigEndian, &headerLen)
	if err != nil {
		log.Printf("读取 HeaderLen 失败: %v", err)
		return
	}
	// 读取包头数据
	headerData := make([]byte, headerLen)
	_, err = io.ReadFull(buf, headerData)
	if err != nil {
		log.Printf("读取包头数据失败: %v", err)
		return
	}
	// 解析包头
	header, err := deserializeHeader(headerData, headerLen)
	if err != nil {
		log.Printf("解析包头失败: %v", err)
		return
	}
	log.Printf("解析包头: %+v", header)
	// 读取数据部分
	dataLen := int(length) - 4 - int(headerLen) // Length(2) + HeaderLen(2) + headerLen + dataLen = total Length
	if dataLen < 0 {
		log.Printf("数据长度计算错误: %d", dataLen)
		return
	}
	data := make([]byte, dataLen)
	_, err = io.ReadFull(buf, data)
	if err != nil {
		log.Printf("读取数据部分失败: %v", err)
		return
	}
	log.Printf("原始 HTTP 请求数据长度: %d bytes", len(data))

	// 解析原始 HTTP 请求
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(data)))
	if err != nil {
		log.Printf("解析 HTTP 请求失败: %v", err)
		return
	}

	// 修改请求目标地址
	req.URL.Scheme = "http"
	req.URL.Host = "47.94.193.70:9090" // 替换为实际目标服务器地址
	req.RequestURI = ""                // 必须清空 RequestURI

	// 创建 HTTP 客户端
	client := &http.Client{}

	// 发送 HTTP 请求
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送 HTTP 请求失败: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("成功发送 HTTP 请求到目标服务器，状态码: %d", resp.StatusCode)

	// 读取 HTTP 响应
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取 HTTP 响应失败: %v", err)
		return
	}
	log.Printf("接收到目标服务器的响应数据长度: %d bytes", len(responseData))

	// 构造 HTTP 响应并返回给代理服务器1
	// 修正状态行，避免重复状态码
	respHeader := fmt.Sprintf("HTTP/1.1 %s\r\n", resp.Status)
	for key, values := range resp.Header {
		for _, value := range values {
			respHeader += fmt.Sprintf("%s: %s\r\n", key, value)
		}
	}
	respHeader += "\r\n"

	// 将响应头和响应体拼接
	fullResponse := []byte(respHeader)
	fullResponse = append(fullResponse, responseData...)

	// 将响应返回给代理服务器1
	_, err = proxy1Conn.Write(fullResponse)
	if err != nil {
		log.Printf("返回响应给代理服务器1失败: %v", err)
		return
	}
	log.Println("响应已返回给代理服务器1")
}

func main() {
	// 启动代理服务器1（HTTP 服务器）
	go startProxyServer1()

	// 启动代理服务器2
	startProxyServer2()
}
