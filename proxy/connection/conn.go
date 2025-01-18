package connection

import (
	"net"
	"sync"
)

// PoolConn 对tcp连接的封装
type PoolConn struct {
	net.Conn
	mu       sync.RWMutex
	c        *channelPool
	unusable bool
}

// Close() 将tcp连接放回池中而非彻底关闭
func (p *PoolConn) Close() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.unusable {
		if p.Conn != nil {
			return p.Conn.Close()
		}
		return nil
	}
	return p.c.put(p.Conn)
}

// MarkUnusable() 用于将连接标记为不可用。这样连接不会被放回池中，而是在 Close() 调用时直接关闭。
func (p *PoolConn) MarkUnusable() {
	p.mu.Lock()
	p.unusable = true
	p.mu.Unlock()
}

// 将一个标准 net.Conn 封装为 PoolConn
func (c *channelPool) wrapConn(conn net.Conn) net.Conn {
	p := &PoolConn{c: c}
	p.Conn = conn
	return p
}
