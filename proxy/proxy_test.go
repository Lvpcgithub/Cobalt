package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProxy(t *testing.T) {
	// 配置文件的加载
	var cfg Config
	// 这里假设config.toml已经存在，你可以根据需要修改
	if _, err := toml.DecodeFile("config.toml", &cfg); err != nil {
		t.Fatalf("Failed to parse config file: %v", err)
	}

	proxy1Config := cfg.Proxy1
	proxy2Config := cfg.Proxy2
	proxy3Config := cfg.Proxy3

	// 启动代理（你可以为每个代理启动一个goroutine）
	go func() {
		startHttpProxy(proxy1Config)
	}()
	go func() {
		startTcpProxy(proxy2Config)
	}()
	go func() {
		startTcpProxy(proxy3Config)
	}()

	// 等待代理启动
	time.Sleep(2 * time.Second)

	// 现在可以模拟一些HTTP请求，来测试代理是否正常工作
	resp, err := http.Get("http://localhost:8081") // 假设Proxy1监听在8081端口
	if err != nil {
		t.Fatalf("Failed to send request to proxy1: %v", err)
	}
	defer resp.Body.Close()

	// 验证响应状态码
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %v", resp.Status)
	}

	// 可选：读取响应体并验证其内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	fmt.Printf("Response body: %s\n", string(body))
}

func TestProxyServer(t *testing.T) {
	// 模拟一个简单的目标服务器，来测试 HTTP 代理的请求转发
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from the target server"))
	}))
	defer targetServer.Close()

	// 设置 Proxy 配置，使其指向测试目标服务器
	cfg := &Config{
		Proxy1: ProxyConfig{
			ListenAddr: ":8081", // 监听地址
			NextProxy:  targetServer.URL,
			ServerAddr: targetServer.URL,
		},
	}

	// 启动 Proxy1
	go startHttpProxy(cfg.Proxy1)

	// 等待 Proxy1 启动
	time.Sleep(1 * time.Second)

	// 发起请求，验证 Proxy1 是否正确转发请求到目标服务器
	resp, err := http.Get("http://localhost:8081")
	if err != nil {
		t.Fatalf("Failed to send request to Proxy1: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %v", resp.Status)
	}

	// 读取并验证响应体内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expected := "Hello from the target server"
	if string(body) != expected {
		t.Fatalf("Expected body %s, got %s", expected, string(body))
	}
}

func TestTcpProxy(t *testing.T) {
	// 这里也可以用类似的方式模拟 TCP 代理的功能
	// 可以创建一个简单的目标服务器并模拟 TCP 流量

	// 启动目标服务器（模拟服务端）
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer ln.Close()

	// 启动 Proxy2
	cfg := &Config{
		Proxy2: ProxyConfig{
			ListenAddr: ":9001",          // 代理监听的端口
			NextProxy:  "localhost:9000", // 目标服务器地址
			ServerAddr: "localhost:9000",
		},
	}
	go startTcpProxy(cfg.Proxy2)

	// 等待 Proxy2 启动
	time.Sleep(1 * time.Second)

	// 模拟 TCP 客户端连接
	conn, err := net.Dial("tcp", ":9001")
	if err != nil {
		t.Fatalf("Failed to connect to Proxy2: %v", err)
	}
	defer conn.Close()

	// 发送简单的请求
	_, err = conn.Write([]byte("Hello Proxy2"))
	if err != nil {
		t.Fatalf("Failed to send data to Proxy2: %v", err)
	}

	// 读取响应
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read data from Proxy2: %v", err)
	}

	// 打印响应内容，验证数据是否正确转发
	fmt.Printf("Received response: %s\n", string(buf[:n]))
}
