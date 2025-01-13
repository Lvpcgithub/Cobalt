package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/BurntSushi/toml"
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
		if config.NextProxy == config.ServerAddr {
			// 如果下一跳是目标服务器则使用 SingleHostReverseProxy 直接转发
			targetURL, err := url.Parse("http://" + config.NextProxy)
			if err != nil {
				log.Printf("解析目标URL失败: %v", err)
				http.Error(w, "内部服务器错误", http.StatusInternalServerError)
				return
			}
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			proxy.ServeHTTP(w, r)
			log.Printf("Proxy: 直接转发请求到 %s", config.NextProxy)
		} else {
			// 否则下一跳仍是代理服务器则使用tcp转发序列化的http请求
			log.Printf("Proxy: 接收到来自客户端的请求: %s %s", r.Method, r.URL)

			// tcp连接到下一跳代理
			conn, err := net.Dial("tcp", config.NextProxy)
			if err != nil {
				log.Printf("Proxy: 连接到下一跳代理 %s 失败: %v", config.NextProxy, err)
				http.Error(w, "无法连接到下一跳代理", http.StatusBadGateway)
				return
			}
			defer conn.Close()

			// 将HTTP请求字节流发送到下一跳代理
			reqBytes, err := httputil.DumpRequest(r, true)
			if err != nil {
				log.Printf("Proxy: 序列化请求失败: %v", err)
				http.Error(w, "无法序列化请求", http.StatusInternalServerError)
				return
			}
			_, err = conn.Write(reqBytes)
			if err != nil {
				log.Printf("Proxy: 发送请求到下一跳代理失败: %v", err)
				http.Error(w, "发送请求失败", http.StatusInternalServerError)
				return
			}
			log.Printf("Proxy: 请求已转发到下一跳代理 %s", config.NextProxy)

			// 从代理读取响应
			reader := bufio.NewReader(conn)
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
	// 注册处理函数
	http.HandleFunc("/", handler)
	log.Printf("Proxy: 正在监听 %s, 转发到 %s", config.ListenAddr, config.NextProxy)
	log.Fatal(http.ListenAndServe(config.ListenAddr, nil))
}

// 处理其他代理的tcp请求
func startTcpProxy(config ProxyConfig) {
	listener, err := net.Listen("tcp", config.ListenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", config.ListenAddr, err)
	}
	defer listener.Close()
	log.Printf("Proxy: Listening on %s", config.ListenAddr)

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

	if config.NextProxy != config.ServerAddr {
		// 如果下一跳是代理服务器则使用tcp原封不动转发接收到的数据
		nextConn, err := net.Dial("tcp", config.NextProxy)
		if err != nil {
			log.Printf("Failed to dial next proxy %s: %v", config.NextProxy, err)
			return
		}
		defer nextConn.Close()
		log.Printf("Proxy: Forwarding connection to %s", config.NextProxy)

		go func() {
			_, err := io.Copy(conn, nextConn)
			if err != nil {
				log.Printf("Proxy: Failed to forward data to client: %v", err)
			}
		}()
		_, err = io.Copy(nextConn, conn)
		if err != nil {
			log.Printf("Proxy: Failed to forward data to next proxy: %v", err)
		}
	} else {
		// 如果下一跳是目标服务器则作为HTTP代理，从tcp收到的数据中读取并处理HTTP请求再转发到目标服务器
		req, err := http.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			log.Printf("Proxy: Failed to read request: %v", err)
			return
		}
		log.Printf("Proxy: Received request: %s %s", req.Method, req.URL)

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
		log.Printf("Proxy: Received response from server: %s", resp.Status)

		respBytes, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Printf("Proxy: Failed to dump response: %v", err)
			return
		}
		_, err = conn.Write(respBytes)
		if err != nil {
			log.Printf("Proxy: Failed to write response to client: %v", err)
		}
	}
}
