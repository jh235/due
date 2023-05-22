package netpoll

import (
	"github.com/cloudwego/netpoll"
	"github.com/dobyte/due/utils/xnet"
	"net"
	"sync"
	"sync/atomic"
)

type serverConn struct {
	rw    sync.RWMutex       // 锁
	id    int64              // 连接ID
	uid   int64              // 用户ID
	state int32              // 连接状态
	conn  netpoll.Connection // 源连接
}

// ID 获取连接ID
func (c *serverConn) ID() int64 {
	return c.id
}

// UID 获取用户ID
func (c *serverConn) UID() int64 {
	return atomic.LoadInt64(&c.uid)
}

// Bind 绑定用户ID
func (c *serverConn) Bind(uid int64) {
	atomic.StoreInt64(&c.uid, uid)
}

// Unbind 解绑用户ID
func (c *serverConn) Unbind() {
	atomic.StoreInt64(&c.uid, 0)
}

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
func (c *serverConn) LocalIP() (string, error) {
	addr, err := c.LocalAddr()
	if err != nil {
		return "", err
	}

	return xnet.ExtractIP(addr)
}

// LocalAddr 获取本地地址
func (c *serverConn) LocalAddr() (net.Addr, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	return c.conn.LocalAddr(), nil
}

// RemoteIP 获取远端IP
func (c *serverConn) RemoteIP() (string, error) {}

// RemoteAddr 获取远端地址
func (c *serverConn) RemoteAddr() (net.Addr, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	return c.conn.RemoteAddr(), nil
}
