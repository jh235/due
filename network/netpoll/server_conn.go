package netpoll

import (
	"github.com/cloudwego/netpoll"
	"net"
)

type serverConn struct {
	conn netpoll.Connection
}

// ID 获取连接ID
func (c *serverConn) ID() int64 {
	//c.conn.Fd()
}

// UID 获取用户ID
func (c *serverConn) UID() int64 {

}

// Bind 绑定用户ID
func (c *serverConn) Bind(uid int64) {}

// Unbind 解绑用户ID
func (c *serverConn) Unbind() {}

// Send 发送消息（同步）
func (c *serverConn) Send(msg []byte, msgType ...int) error {}

// Push 发送消息（异步）
func (c *serverConn) Push(msg []byte, msgType ...int) error {}

// State 获取连接状态
func (c *serverConn) State() ConnState {}

// Close 关闭连接
func (c *serverConn) Close(isForce ...bool) error {
	return c.conn.Close()
}

// LocalIP 获取本地IP
func (c *serverConn) LocalIP() (string, error) {}

// LocalAddr 获取本地地址
func (c *serverConn) LocalAddr() (net.Addr, error) {}

// RemoteIP 获取远端IP
func (c *serverConn) RemoteIP() (string, error) {}

// RemoteAddr 获取远端地址
func (c *serverConn) RemoteAddr() (net.Addr, error) {}
