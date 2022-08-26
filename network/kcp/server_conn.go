package kcp

import (
	"fmt"
	"github.com/dobyte/due/internal/xnet"
	"github.com/dobyte/due/log"
	"github.com/dobyte/due/network"
	"net"
	"sync"
	"sync/atomic"
)

type serverConn struct {
	rw      sync.RWMutex   // 锁
	id      int64          // 连接ID
	uid     int64          // 用户ID
	state   int32          // 连接状态
	conn    net.Conn       // WS源连接
	connMgr *serverConnMgr // 连接管理
	chWrite chan chWrite   // 写入队列
	done    chan struct{}  // 写入完成信号
}

var _ network.Conn = &serverConn{}

// ID 获取连接ID
func (c *serverConn) ID() int64 {
	return c.id
}

// UID 获取用户ID
func (c *serverConn) UID() int64 {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return c.uid
}

// Bind 绑定用户ID
func (c *serverConn) Bind(uid int64) {
	c.rw.Lock()
	defer c.rw.Unlock()

	c.uid = uid
}

// Send 发送消息（同步）
func (c *serverConn) Send(msg []byte, msgType ...int) error {
	c.rw.RLock()
	defer c.rw.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	_, err := c.conn.Write(msg)

	return err
}

// Push 发送消息（异步）
func (c *serverConn) Push(msg []byte, msgType ...int) error {
	c.rw.RLock()
	defer c.rw.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	c.chWrite <- chWrite{typ: dataPacket, msg: msg}

	return nil
}

// State 获取连接状态
func (c *serverConn) State() network.ConnState {
	return network.ConnState(atomic.LoadInt32(&c.state))
}

// Close 关闭连接
func (c *serverConn) Close(isForce ...bool) error {
	c.rw.Lock()
	defer c.rw.Unlock()

	if err := c.checkState(); err != nil {
		return err
	}

	if len(isForce) > 0 && isForce[0] {
		atomic.StoreInt32(&c.state, int32(network.ConnClosed))
	} else {
		atomic.StoreInt32(&c.state, int32(network.ConnHanged))
		c.chWrite <- chWrite{typ: closeSig}
		<-c.done
	}

	close(c.chWrite)

	err := c.conn.Close()
	c.conn = nil
	c.connMgr.recycle(c)

	return err
}

// 关闭连接
func (c *serverConn) close() {
	atomic.StoreInt32(&c.state, int32(network.ConnClosed))

	if c.connMgr.server.disconnectHandler != nil {
		c.connMgr.server.disconnectHandler(c)
	}
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
func (c *serverConn) RemoteIP() (string, error) {
	addr, err := c.RemoteAddr()
	if err != nil {
		return "", err
	}

	return xnet.ExtractIP(addr)
}

// RemoteAddr 获取远端地址
func (c *serverConn) RemoteAddr() (net.Addr, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	return c.conn.RemoteAddr(), nil
}

// 初始化连接
func (c *serverConn) init(conn net.Conn, cm *serverConnMgr) {
	c.id = cm.id
	c.conn = conn
	c.connMgr = cm
	c.chWrite = make(chan chWrite, 256)
	c.done = make(chan struct{})
	atomic.StoreInt32(&c.state, int32(network.ConnOpened))

	if c.connMgr.server.connectHandler != nil {
		c.connMgr.server.connectHandler(c)
	}

	go c.read()

	go c.write()
}

// 检测连接状态
func (c *serverConn) checkState() error {
	switch network.ConnState(atomic.LoadInt32(&c.state)) {
	case network.ConnHanged:
		return network.ErrConnectionHanged
	case network.ConnClosed:
		return network.ErrConnectionClosed
	}

	return nil
}

// 读取消息
func (c *serverConn) read() {
	defer c.close()

	size := c.connMgr.server.opts.maxMsgLength + 1
	buf := make([]byte, 9)

	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			fmt.Println("3333")
			return
		}

		fmt.Println(n)

		if n >= size {
			log.Warnf("the msg size too large, has been ignored")
			continue
		}

		switch c.State() {
		case network.ConnHanged:
			continue
		case network.ConnClosed:
			return
		}

		if c.connMgr.server.receiveHandler != nil {
			c.connMgr.server.receiveHandler(c, buf[:n], 0)
		}
	}
}

// 写入消息
func (c *serverConn) write() {
	for {
		select {
		case write, ok := <-c.chWrite:
			if !ok {
				return
			}

			if write.typ == closeSig {
				c.done <- struct{}{}
				return
			}

			if _, err := c.conn.Write(write.msg); err != nil {
				log.Errorf("write message error: %v", err)
			}
		}
	}
}
