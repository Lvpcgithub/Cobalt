package main

import (
	"demo1/tcp"
	"sync"

	"demo1/info"
	"demo1/tcp_probe"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(3)

	// 启动 API 服务
	go func() {
		defer wg.Done()
		tcp_probe.StartAPIServer()
	}()

	// 启动信息收集和发送的模块
	go func() {
		defer wg.Done()
		info.StartInfoCollector()
	}()

	go func() {
		defer wg.Done()
		tcp.StartTCPServer("50000")
	}()

	wg.Wait()
}
