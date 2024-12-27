package handler

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTP 客户端（复用连接）
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	},
}

// Module1API: 模块1的对外接口
type Module1API struct {
	ProxyNodeAPI *Module2API // 模块 2 的接口实例
}

// NewModule1API: 创建模块1实例
func NewModule1API(proxyNodeAPI *Module2API) *Module1API {
	return &Module1API{ProxyNodeAPI: proxyNodeAPI}
}

// StartClientServer: 启动HTTP服务器监听客户端请求
func (api *Module1API) StartClientServer(addr string) error {
	http.HandleFunc("/", api.handleClientRequest)
	fmt.Printf("ClientServer: Listening on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// handleClientRequest: 处理来自客户端的HTTP请求
func (api *Module1API) handleClientRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received request: %s %s\n", r.Method, r.URL.String())

	// 确定下一跳
	nextHop, isProxy := api.determineNextHop(r)
	if nextHop == "" {
		http.Error(w, "No next hop found", http.StatusBadRequest)
		return
	}

	if isProxy {
		// 如果下一跳是代理节点，交给模块2处理
		responseBytes, err := api.forwardToProxy(nextHop, r)
		if err != nil {
			http.Error(w, "Failed to forward request to proxy", http.StatusInternalServerError)
			return
		}

		// 返回模块2的响应给客户端
		w.Write(responseBytes)
	} else {
		// 如果下一跳是目标服务器，直接处理
		resp, err := api.forwardToServer(r, nextHop)
		if err != nil {
			http.Error(w, "Failed to forward request to server", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close() // 确保响应体关闭以释放资源

		// 设置状态码
		w.WriteHeader(resp.StatusCode)

		// 将响应体数据写入客户端响应
		_, copyErr := io.Copy(w, resp.Body)
		if copyErr != nil {
			fmt.Printf("Failed to copy response body: %v\n", copyErr)
		}
	}
}

// determineNextHop: 确定下一跳（代理节点或目标服务器）
func (api *Module1API) determineNextHop(r *http.Request) (string, bool) {
	// 示例实现：根据路径判断下一跳
	if r.URL.Path == "/proxy" {
		return "localhost:9000", true // 返回代理节点地址
	}
	return "http://example.com", false // 返回目标服务器地址
}

// forwardToProxy: 将请求转发到代理节点（模块2）
func (api *Module1API) forwardToProxy(proxyAddr string, r *http.Request) ([]byte, error) {
	// 将请求转换为适合模块2的格式
	headers := make(map[string][]string)
	for key, values := range r.Header {
		headers[key] = values
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// 调用模块2的接口
	return api.ProxyNodeAPI.SendRequestToProxy(proxyAddr, r.Method, r.URL.String(), headers, body)
}

// forwardToServer: 转发HTTP请求到目标服务器
func (api *Module1API) forwardToServer(r *http.Request, targetURL string) (*http.Response, error) {
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header = r.Header
	return httpClient.Do(req)
}
