/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2022/7/7 1:19 上午
 * @Desc: 网关服务器
 */

package gate

import (
	"context"
	"sync"
	"time"

	"github.com/dobyte/due/cluster"
	"github.com/dobyte/due/cluster/internal/pb"
	"github.com/dobyte/due/internal/xnet"
	"github.com/dobyte/due/packet"
	"github.com/dobyte/due/registry"
	"github.com/dobyte/due/router"
	"github.com/dobyte/due/session"

	"github.com/google/uuid"

	"github.com/dobyte/due/component"
	"github.com/dobyte/due/log"
	"github.com/dobyte/due/network"
)

const (
	defaultName    = "gate"
	defaultTimeout = 3 * time.Second // 默认超时时间
)

type Gate struct {
	component.Base
	opts     *options
	ctx      context.Context
	cancel   context.CancelFunc
	group    *session.Group
	sessions sync.Pool
	proxy    *proxy
	router   *router.Router
	instance *registry.ServiceInstance
}

func NewGate(opts ...Option) *Gate {
	o := &options{
		ctx:     context.Background(),
		name:    defaultName,
		timeout: defaultTimeout,
	}
	if id, err := uuid.NewUUID(); err == nil {
		o.id = id.String()
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.id == "" {
		log.Fatal("the gate instance ID is not registered")
	}
	if o.redis == nil {
		log.Fatal("the redis client is not registered.")
	}
	if o.server == nil {
		log.Fatal("the gate server is not registered.")
	}
	if o.grpc == nil {
		log.Fatal("the grpc server is not registered.")
	}
	if o.registry == nil {
		log.Fatal("the registry is not registered.")
	}

	g := &Gate{}
	g.opts = o
	g.group = session.NewGroup()
	g.proxy = newProxy(g)
	g.router = router.NewRouter()
	g.sessions.New = func() interface{} { return session.NewSession() }
	g.ctx, g.cancel = context.WithCancel(o.ctx)

	return g
}

// Name 组件名称
func (g *Gate) Name() string {
	return g.opts.name
}

// Init 初始化
func (g *Gate) Init() {
	g.buildInstance()
}

// Start 启动组件
func (g *Gate) Start() {
	g.startGate()

	g.startGRPC()

	g.startProxy()

	g.registry()

	g.debugPrint()
}

// Destroy 销毁组件
func (g *Gate) Destroy() {
	g.stopGate()

	g.stopGRPC()

	g.cancel()
}

// 启动Gate服务
func (g *Gate) startGate() {
	g.opts.server.OnConnect(g.handleConnect)
	g.opts.server.OnDisconnect(g.handleDisconnect)
	g.opts.server.OnReceive(g.handleReceive)

	if err := g.opts.server.Start(); err != nil {
		log.Fatalf("the gate server startup failed: %v", err)
	}
}

// 停止Gate服务
func (g *Gate) stopGate() {
	if err := g.opts.server.Stop(); err != nil {
		log.Errorf("the gate server stop failed: %v", err)
	}
}

// 启动GRPC服务
func (g *Gate) startGRPC() {
	go func() {
		g.opts.grpc.RegisterService(&pb.Gate_ServiceDesc, &endpoint{gate: g})
		if err := g.opts.grpc.Start(); err != nil {
			log.Fatalf("the grpc server startup failed: %v", err)
		}
	}()
}

// 停止GRPC服务
func (g *Gate) stopGRPC() {
	if err := g.opts.registry.Deregister(g.instance); err != nil {
		log.Errorf("the gate service instance deregister failed: %v", err)
	}

	if err := g.opts.grpc.Stop(); err != nil {
		log.Errorf("the grpc server stop failed: %v", err)
	}
}

// 启动实例代理
func (g *Gate) startProxy() {
	go g.proxy.listen(g.ctx)
}

// 处理连接打开
func (g *Gate) handleConnect(conn network.Conn) {
	s := g.sessions.Get().(*session.Session)
	s.Init(conn)
	g.group.AddSession(s)
}

// 处理断开连接
func (g *Gate) handleDisconnect(conn network.Conn) {
	s, err := g.group.RemSession(session.Conn, conn.ID())
	if err != nil {
		log.Errorf("session remove failed, gid: %d, cid: %d, uid: %d, err: %v", g.opts.id, s.CID(), s.UID(), err)
		return
	}

	if uid := conn.UID(); uid > 0 {
		ctx, cancel := context.WithTimeout(g.ctx, g.opts.timeout)
		err = g.proxy.unbindGate(ctx, uid)
		cancel()
		if err != nil {
			log.Errorf("user unbind failed, gid: %d, uid: %d, err: %v", g.opts.id, uid, err)
		}
	}

	s.Reset()
	g.sessions.Put(s)
}

// 处理接收到的消息
func (g *Gate) handleReceive(conn network.Conn, data []byte, _ int) {
	message, err := packet.Unpack(data)
	if err != nil {
		log.Errorf("unpack data to struct failed: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(g.ctx, g.opts.timeout)
	if err = g.proxy.deliver(ctx, conn.ID(), conn.UID(), message); err != nil {
		log.Errorf("deliver message failed: %v", err)
	}
	cancel()
}

// 注册服务实例
func (g *Gate) registry() {
	if err := g.opts.registry.Register(g.instance); err != nil {
		log.Fatalf("the gate service instance register failed: %v", err)
	}

	watcher, err := g.opts.registry.Watch(context.Background(), string(cluster.Node))
	if err != nil {
		log.Fatalf("the node service watch failed: %v", err)
	}

	go func() {
		for {
			services, err := watcher.Next()
			if err != nil {
				continue
			}
			g.router.ReplaceServices(services...)
		}
	}()
}

// 构建服务实例
func (g *Gate) buildInstance() {
	g.instance = &registry.ServiceInstance{
		ID:       g.opts.id,
		Name:     string(cluster.Gate),
		Endpoint: g.opts.grpc.Endpoint().String(),
	}
}

func (g *Gate) debugPrint() {
	log.Debugf("The gate server startup successful")
	log.Debugf("Gate server, listen: %s protocol: %s", xnet.FulfillAddr(g.opts.server.Addr()), g.opts.server.Protocol())
	log.Debugf("GRPC server, listen: %s protocol: %s", xnet.FulfillAddr(g.opts.grpc.Addr()), g.opts.grpc.Scheme())
}
