package connection

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

// channelPool 基于缓冲通道实现pool接口
type channelPool struct {
	// 存储连接
	mu    sync.RWMutex
	conns chan net.Conn

	// 生成连接
	factory Factory
}

// Factory 生成连接的函数
type Factory func() (net.Conn, error)

// NewChannelPool 返回一个基于缓冲通道的新池，参数为初始值容量、最大容量和Factory函数。
// Factory函数在初始容量大于零时填充连接池。
// Get()时，如果没有新连接在池中可用，将通过Factory函数创建一个新连接
func NewChannelPool(initialCap, maxCap int, factory Factory) (Pool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}

	c := &channelPool{
		conns:   make(chan net.Conn, maxCap),
		factory: factory,
	}

	// 创建初始容量的连接
	for i := 0; i < initialCap; i++ {
		conn, err := factory()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		c.conns <- conn
	}
	return c, nil
}

// getConnsAndFactory 用于安全地获取连接通道和工厂函数，采用读锁保护数据的并发访问
func (c *channelPool) getConnsAndFactory() (chan net.Conn, Factory) {
	c.mu.RLock()
	conns := c.conns
	factory := c.factory
	c.mu.RUnlock()
	return conns, factory
}

// Get 实现 Pool 接口 Get() 方法.
// 如果连接池中没有可用连接，调用factory函数创建新连接
func (c *channelPool) Get() (net.Conn, error) {
	conns, factory := c.getConnsAndFactory()
	if conns == nil {
		return nil, ErrClosed
	}
	// 返回 wrapConn(conn)，即封装后的连接，以便调用 Close() 时可以返回池中。
	select {
	case conn := <-conns:
		if conn == nil {
			return nil, ErrClosed
		}

		return c.wrapConn(conn), nil
	default:
		conn, err := factory()
		if err != nil {
			return nil, err
		}

		return c.wrapConn(conn), nil
	}
}

// put 将连接放回连接池. 连接池满了或者关闭了则彻底关闭连接,
func (c *channelPool) put(conn net.Conn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// 连接池如果关闭则彻底关闭连接
	if c.conns == nil {
		return conn.Close()
	}

	// 若池未满，将连接放回池中；若池已满，则直接关闭连接。
	select {
	case c.conns <- conn:
		return nil
	default:
		return conn.Close()
	}
}

func (c *channelPool) Close() {
	c.mu.Lock()
	conns := c.conns
	c.conns = nil
	c.factory = nil
	c.mu.Unlock()

	if conns == nil {
		return
	}

	close(conns)
	for conn := range conns {
		conn.Close()
	}
}

func (c *channelPool) Len() int {
	conns, _ := c.getConnsAndFactory()
	return len(conns)
}
