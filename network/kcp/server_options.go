/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2022/7/12 12:57 上午
 * @Desc: TODO
 */

package kcp

type ServerOption func(o *serverOptions)

type serverOptions struct {
	addr         string // 监听地址
	maxConnNum   int    // 最大连接数
	maxMsgLength int    // 最大消息长度
}

// WithServerListenAddr 设置监听地址
func WithServerListenAddr(addr string) ServerOption {
	return func(o *serverOptions) { o.addr = addr }
}

// WithServerMaxConnNum 设置连接的最大连接数
func WithServerMaxConnNum(maxConnNum int) ServerOption {
	return func(o *serverOptions) { o.maxConnNum = maxConnNum }
}

// WithServerMaxMsgLength 设置消息最大长度
func WithServerMaxMsgLength(maxMsgLength int) ServerOption {
	return func(o *serverOptions) { o.maxMsgLength = maxMsgLength }
}
