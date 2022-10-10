package node

import (
	"context"
	"github.com/dobyte/due/encoding"
	"github.com/dobyte/due/locate"
	"github.com/dobyte/due/registry"
	"github.com/dobyte/due/transport/grpc"
	"time"
)

type Option func(o *options)

type options struct {
	id       string            // 实例ID
	name     string            // 实例名称
	ctx      context.Context   // 上下文
	grpc     *grpc.Server      // GRPC服务器
	codec    encoding.Codec    // 编解码器
	timeout  time.Duration     // RPC调用超时时间
	locator  locate.Locator    // 定位器
	registry registry.Registry // 服务注册
}

// WithID 设置实例ID
func WithID(id string) Option {
	return func(o *options) { o.id = id }
}

// WithName 设置实例名称
func WithName(name string) Option {
	return func(o *options) { o.name = name }
}

// WithCodec 设置编解码器
func WithCodec(codec encoding.Codec) Option {
	return func(o *options) { o.codec = codec }
}

// WithContext 设置上下文
func WithContext(ctx context.Context) Option {
	return func(o *options) { o.ctx = ctx }
}

// WithGRPCServer 设置GRPC服务器
func WithGRPCServer(grpc *grpc.Server) Option {
	return func(o *options) { o.grpc = grpc }
}

// WithTimeout 设置RPC调用超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) { o.timeout = timeout }
}

// WithLocator 设置定位器
func WithLocator(locator locate.Locator) Option {
	return func(o *options) { o.locator = locator }
}

// WithRegistry 设置服务注册
func WithRegistry(r registry.Registry) Option {
	return func(o *options) { o.registry = r }
}
