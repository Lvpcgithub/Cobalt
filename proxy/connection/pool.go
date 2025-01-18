package connection

import (
	"errors"
	"net"
)

var (
	// ErrClosed 连接池关闭时调用pool.Close()错误
	ErrClosed = errors.New("pool is closed")
)

// Pool 接口
type Pool interface {
	// Get 返回从连接池获取的一个tcp连接
	Get() (net.Conn, error)

	// Close 关闭连接池以及所有连接
	Close()

	// Len 返回连接池中连接数量
	Len() int
}
