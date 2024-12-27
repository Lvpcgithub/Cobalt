package smux_usage

import (
	"github.com/xtaci/smux"
	"log"
	"net"
	"time"
)

// smuxSessions 用来维护 TCP 连接与 SMUX 会话之间的映射关系
var smuxSessions = make(map[net.Conn]*smux.Session)

// GetOrCreateSMUXSession 创建或获取SMUX会话
func GetOrCreateSMUXSession(conn net.Conn) (*smux.Session, error) {
	// 检查是否已有与该连接相关的SMUX会话
	if session, exists := smuxSessions[conn]; exists {
		return session, nil
	}

	// 如果SMUX会话不存在，则为该TCP连接创建新的会话
	session, err := CreateSMUXSession(conn)
	if err != nil {
		log.Printf("Failed to create SMUX session: %v", err)
		return nil, err
	}

	// 将SMUX会话保存到映射中
	smuxSessions[conn] = session
	log.Println("SMUX session created successfully")
	return session, nil
}

// CloseSMUXSessionByConn 关闭与给定 TCP 连接相关的 SMUX 会话
func CloseSMUXSessionByConn(conn net.Conn) {
	if session, exists := smuxSessions[conn]; exists {
		session.Close()
		delete(smuxSessions, conn) // 删除映射，释放资源
		log.Println("SMUX session closed and removed from session map")
	}
}

// CreateSMUXSession 创建一个 SMUX 客户端会话，通过给定的 TCP 连接
func CreateSMUXSession(conn net.Conn) (*smux.Session, error) {
	// 创建 SMUX 配置，这里可以设置超时、流控制等参数
	config := smux.DefaultConfig()
	config.KeepAliveTimeout = 10 * time.Second // 配置保持活动超时

	// 通过 TCP 连接创建一个 SMUX 会话
	session, err := smux.Client(conn, config)
	if err != nil {
		log.Printf("Failed to create SMUX session: %v", err)
		return nil, err
	}

	log.Println("SMUX session created successfully")
	return session, nil
}

// AcceptSMUXSession 用于在代理 B 或服务端接受 SMUX 连接
func AcceptSMUXSession(conn net.Conn) (*smux.Session, error) {
	// 使用默认配置创建 SMUX 会话
	config := smux.DefaultConfig()

	// 在已有的 TCP 连接上创建一个 SMUX 服务端会话
	session, err := smux.Server(conn, config)
	if err != nil {
		log.Printf("Failed to accept SMUX session: %v", err)
		return nil, err
	}

	log.Println("SMUX session accepted successfully")
	return session, nil
}

// OpenSMUXStream 用于在一个现有的 SMUX 会话中打开一个新的流
func OpenSMUXStream(session *smux.Session) (*smux.Stream, error) {
	// 打开一个新的 SMUX 流
	stream, err := session.OpenStream()
	if err != nil {
		log.Printf("Failed to open SMUX stream: %v", err)
		return nil, err
	}

	log.Println("SMUX stream opened successfully")
	return stream, nil
}

// AcceptSMUXStream 用于在代理 B 或服务端接受新的 SMUX 流
func AcceptSMUXStream(session *smux.Session) (*smux.Stream, error) {
	// 接受一个 SMUX 流
	stream, err := session.AcceptStream()
	if err != nil {
		log.Printf("Failed to accept SMUX stream: %v", err)
		return nil, err
	}

	log.Println("SMUX stream accepted successfully")
	return stream, nil
}
