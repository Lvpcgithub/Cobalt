package handler

import (
	"fmt"
	"log"
)

func main() {
	// 创建模块1和模块2的实例
	var module1 *Module1API
	var module2 *Module2API

	// 创建模块2（依赖模块1）
	module2 = NewModule2API(module1)

	// 创建模块1（依赖模块2）
	module1 = NewModule1API(module2)

	// 更新模块2中的模块1依赖（避免循环引用问题）
	module2.ClientServerAPI = module1

	// 启动模块2（代理节点服务）
	go func() {
		err := module2.StartProxyServer(":9000") // 模块2监听9000端口
		if err != nil {
			log.Fatalf("Failed to start Module2 (ProxyNode): %v", err)
		}
	}()

	// 启动模块1（HTTP服务）
	go func() {
		err := module1.StartClientServer(":8080") // 模块1监听8080端口
		if err != nil {
			log.Fatalf("Failed to start Module1 (ClientServer): %v", err)
		}
	}()

	// 保持主程序运行
	fmt.Println("Both Module1 and Module2 are running...")
	select {}
}
