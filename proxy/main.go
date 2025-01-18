package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/BurntSushi/toml"
	"github.com/xtaci/smux"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"proxy-demo/connection"
	"sync"
	"time"
)

type Config struct {
	Proxy1 ProxyConfig
	Proxy2 ProxyConfig
	Proxy3 ProxyConfig
}

type ProxyConfig struct {
	ListenAddr string // 本地监听地址
	NextProxy  string // 下一跳代理的地址（对于proxy1和proxy2）
	ServerAddr string // 服务器地址（对于proxy3）
}

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

func createConnectionFactory(addr string) connection.Factory {
	return func() (net.Conn, error) {
		return net.Dial("tcp", addr)
	}
}

func main() {
	var cfg Config
	if _, err := toml.DecodeFile("config.toml", &cfg); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	proxy1Config := cfg.Proxy1
	proxy2Config := cfg.Proxy2
	proxy3Config := cfg.Proxy3

	go startHttpProxy(proxy1Config)
	go startTcpProxy(proxy2Config)
	startTcpProxy(proxy3Config)
}

// 处理客户端http请求
func startHttpProxy(config ProxyConfig) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Proxy1: 接收到客户端请求 %s %s", r.Method, r.URL)
		if config.NextProxy == config.ServerAddr {
			// 如果下一跳是目标服务器则使用 SingleHostReverseProxy 直接转发
			targetURL, err := url.Parse("http://" + config.NextProxy)
			if err != nil {
				log.Printf("Proxy1: 解析目标URL失败: %v", err)
				http.Error(w, "内部服务器错误", http.StatusInternalServerError)
				return
			}
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			proxy.ServeHTTP(w, r)
			log.Printf("Proxy1: 请求已转发到目标服务器 %s", config.ServerAddr)
		} else {
			// 否则下一跳仍是代理服务器则使用smux转发序列化的http请求

			// 使用闭包创建工厂函数
			factory := createConnectionFactory(config.NextProxy)

			pool, err := connection.NewChannelPool(5, 10, factory)
			if err != nil {
				log.Fatal(err)
			}
			defer pool.Close()

			// 从连接池获取连接
			conn, err := pool.Get()
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()
			// 创建 smux 会话
			session, err := smux.Client(conn, nil)
			if err != nil {
				log.Printf("Proxy: Failed to create smux session: %v", err)
				http.Error(w, "无法建立与下一跳代理的smux会话", http.StatusBadGateway)
				return
			}
			defer session.Close()

			log.Printf("Proxy1: 已通过smux转发请求到代理2")

			// 为当前请求创建一个新的 smux 流
			stream, err := session.OpenStream()
			if err != nil {
				log.Printf("Proxy: Failed to open stream to next proxy: %v", err)
				http.Error(w, "无法打开流", http.StatusInternalServerError)
				return
			}
			defer stream.Close()
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
			// 将HTTP请求序列化
			reqBytes, err := httputil.DumpRequest(r, true)
			if err != nil {
				log.Printf("Proxy: 序列化请求失败: %v", err)
				http.Error(w, "无法序列化请求", http.StatusInternalServerError)
				return
			}

			// 使用 serializeHeader 将包头与请求一起序列化
			data, err := serializeHeader(header, reqBytes)
			if err != nil {
				log.Printf("Proxy: 序列化包头失败: %v", err)
				http.Error(w, "无法序列化包头", http.StatusInternalServerError)
				return
			}
			_, err = stream.Write(data)
			if err != nil {
				log.Printf("Proxy: 发送请求到下一跳代理失败: %v", err)
				http.Error(w, "发送请求失败", http.StatusInternalServerError)
				return
			}
			log.Printf("Proxy: 请求已转发到下一跳代理 %s", config.NextProxy)

			// 从下一跳代理读取响应
			reader := bufio.NewReader(stream)
			resp, err := http.ReadResponse(reader, r)
			if err != nil {
				log.Printf("Proxy: 从下一跳代理读取响应失败: %v", err)
				http.Error(w, "无法读取响应", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			// 将响应头写回客户端
			for key, values := range resp.Header {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}
			w.WriteHeader(resp.StatusCode)

			// 将响应体写回客户端
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				log.Printf("Proxy: 将响应写回客户端失败: %v", err)
			}
		}
	}
	http.HandleFunc("/", handler)
	log.Printf("Proxy1: 正在监听 %s, 转发到 %s", config.ListenAddr, config.NextProxy)
	log.Fatal(http.ListenAndServe(config.ListenAddr, nil))
}

// 处理其他代理的tcp请求
func startTcpProxy(config ProxyConfig) {
	listener, err := net.Listen("tcp", config.ListenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", config.ListenAddr, err)
	}
	defer listener.Close()
	log.Printf("Proxy2: 正在监听 %s", config.ListenAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleTcpProxyConnection(conn, config)
	}
}

func handleTcpProxyConnection(conn net.Conn, config ProxyConfig) {
	defer conn.Close()

	// 创建 smux 会话
	session, err := smux.Server(conn, nil)
	if err != nil {
		log.Printf("Proxy: Failed to create smux session: %v", err)
		return
	}
	defer session.Close()

	log.Printf("Proxy2: 已接收到smux连接")

	// 创建一个新的 stream，用于处理当前连接的数据
	stream, err := session.AcceptStream()
	if err != nil {
		log.Printf("Proxy: Failed to accept stream: %v", err)
		return
	}
	defer stream.Close()

	if config.NextProxy != config.ServerAddr {
		// 如果下一跳是代理服务器则使用tcp原封不动转发接收到的数据
		log.Printf("Proxy2: 通过smux转发请求到代理3")
		// 使用闭包创建工厂函数
		factory := createConnectionFactory(config.NextProxy)

		pool, err := connection.NewChannelPool(5, 10, factory)
		if err != nil {
			log.Fatal(err)
		}
		defer pool.Close()

		// 从连接池获取连接
		nextConn, err := pool.Get()
		if err != nil {
			log.Fatal(err)
		}
		defer nextConn.Close()
		log.Printf("Proxy: Forwarding connection to %s", config.NextProxy)

		// 在下一跳代理上创建 smux 会话作为客户端
		nextSession, err := smux.Client(nextConn, nil)
		if err != nil {
			log.Printf("Proxy: Failed to create smux client session: %v", err)
			return
		}
		defer nextSession.Close()
		log.Printf("Proxy2: 已通过smux转发请求到代理3")

		// 为当前流创建一个新的流（发送给下一跳代理）
		nextStream, err := nextSession.OpenStream()
		if err != nil {
			log.Printf("Proxy: Failed to open stream to next proxy: %v", err)
			return
		}
		defer nextStream.Close()

		// 使用同步原语保证并发数据转发
		var wg sync.WaitGroup
		wg.Add(2)

		// 转发数据：客户端流 -> 代理3流
		go func() {
			defer wg.Done()
			_, err := io.Copy(nextStream, stream)
			if err != nil && err.Error() != "EOF" {
				log.Printf("Proxy2: 转发数据到代理3失败: %v", err)
			}
		}()

		// 转发数据：代理3流 -> 客户端流
		go func() {
			defer wg.Done()
			_, err := io.Copy(stream, nextStream)
			if err != nil && err.Error() != "EOF" {
				log.Printf("Proxy2: 转发数据从代理3到客户端失败: %v", err)
			}
		}()

		// 等待两个并发操作完成
		wg.Wait()
	} else {
		// 如果下一跳是目标服务器则作为HTTP代理，从tcp收到的数据中读取并处理HTTP请求再转发到目标服务器
		log.Printf("Proxy2: 将请求转发到目标服务器")

		// 读取包头长度部分 (Length + HeaderLen)
		headerLenBytes := make([]byte, 4) // Length(2) + HeaderLen(2)
		_, err = io.ReadFull(stream, headerLenBytes)
		if err != nil {
			log.Printf("Proxy2: Failed to read header length: %v", err)
			return
		}

		// 获取包头长度信息
		length := binary.BigEndian.Uint16(headerLenBytes[:2])    // 包总长度
		headerLen := binary.BigEndian.Uint16(headerLenBytes[2:]) // 包头长度

		// 读取包头数据并解析
		headerData := make([]byte, headerLen)
		_, err = io.ReadFull(stream, headerData)
		if err != nil {
			log.Printf("Proxy2: Failed to read header data: %v", err)
			return
		}

		// 解析自定义包头
		header, err := deserializeHeader(headerData, headerLen)
		if err != nil {
			log.Printf("Proxy2: Failed to deserialize header: %v", err)
			return
		}

		// 输出解析的包头信息（可选）
		log.Printf("Proxy2: 解析包头：Timestamp=%d, PacketID=%d, PacketType=%d", header.Timestamp, header.PacketID, header.PacketType)

		// 读取 HTTP 请求数据
		reqBytes := make([]byte, length-uint16(2+2+headerLen)) // 总长度减去 Length(2) 和 HeaderLen(2) 部分
		_, err = io.ReadFull(stream, reqBytes)
		if err != nil {
			log.Printf("Proxy2: Failed to read request body: %v", err)
			return
		}

		// 反序列化 HTTP 请求
		req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(reqBytes)))
		if err != nil {
			log.Printf("Proxy: Failed to read HTTP request: %v", err)
			return
		}

		// 将 HTTP 请求转发到目标服务器
		log.Printf("Proxy2: 将请求转发到目标服务器")
		newReq, err := http.NewRequest(req.Method, "http://"+config.ServerAddr+req.URL.Path, req.Body)
		if err != nil {
			log.Printf("Proxy: Failed to create new request: %v", err)
			return
		}
		for key, values := range req.Header {
			for _, value := range values {
				newReq.Header.Add(key, value)
			}
		}

		client := &http.Client{}
		resp, err := client.Do(newReq)
		if err != nil {
			log.Printf("Proxy: Failed to send request to server: %v", err)
			return
		}
		defer resp.Body.Close()

		// 输出目标服务器的响应状态（可选）
		log.Printf("Proxy2: 从目标服务器收到响应: %s", resp.Status)

		// 直接将目标服务器的响应转发给客户端
		respBytes, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Printf("Proxy: Failed to dump response: %v", err)
			return
		}

		_, err = stream.Write(respBytes)
		if err != nil {
			log.Printf("Proxy: Failed to write response to client: %v", err)
		}
	}
}
