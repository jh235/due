package transport

import (
	"context"
	"github.com/dobyte/due/internal/endpoint"
	"github.com/dobyte/due/router"
)

type Transporter interface {
	// NewServer 新建服务器
	NewServer() (Server, error)
	// NewClient 新建客户端
	NewClient(ep *router.Endpoint) (Client, error)

	// NewGateServer 新建网关服务器
	NewGateServer(provider GateProvider) (Server, error)
	// NewNodeServer 新建节点服务器
	NewNodeServer(provider NodeProvider) (Server, error)
	// NewGateClient 新建网关客户端
	NewGateClient(ep *router.Endpoint) (GateClient, error)
	// NewNodeClient 新建节点客户端
	NewNodeClient(ep *router.Endpoint) (NodeClient, error)
}

type Server interface {
	// Addr 监听地址
	Addr() string
	// Scheme 协议
	Scheme() string
	// Endpoint 服务端口
	Endpoint() *endpoint.Endpoint
	// Start 启动服务器
	Start() error
	// Stop 停止服务器
	Stop() error
}

type Client interface {
	// Invoke 调用方法
	Invoke(ctx context.Context, method string, args, reply interface{}) error
}
