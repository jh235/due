package node

import (
	"context"
	"github.com/dobyte/due/transport"
	"sync"
	"time"

	"github.com/dobyte/due/cluster"
	"github.com/dobyte/due/internal/xnet"
	"github.com/dobyte/due/registry"
	"github.com/dobyte/due/router"

	"github.com/dobyte/due/component"
	"github.com/dobyte/due/log"
)

type RouteHandler func(req Request)

type EventHandler func(gid string, uid int64)

type routeEntity struct {
	route    int32        // 路由
	stateful bool         // 是否有状态
	handler  RouteHandler // 路由处理器
}

type eventEntity struct {
	event cluster.Event
	gid   string
	uid   int64
}

type Node struct {
	component.Base
	opts                *options
	ctx                 context.Context
	cancel              context.CancelFunc
	chEvent             chan *eventEntity
	chRequest           chan *request
	rw                  sync.RWMutex
	routes              map[int32]routeEntity
	defaultRouteHandler RouteHandler
	events              map[cluster.Event]EventHandler
	proxy               *proxy
	router              *router.Router
	instance            *registry.ServiceInstance
	rpc                 transport.Server
}

func NewNode(opts ...Option) *Node {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	n := &Node{}
	n.opts = o
	n.routes = make(map[int32]routeEntity)
	n.events = make(map[cluster.Event]EventHandler)
	n.chEvent = make(chan *eventEntity, 1024)
	n.chRequest = make(chan *request, 1024)
	n.proxy = newProxy(n)
	n.router = router.NewRouter()
	n.ctx, n.cancel = context.WithCancel(o.ctx)

	return n
}

// Name 组件名称
func (n *Node) Name() string {
	return n.opts.name
}

// Init 初始化节点
func (n *Node) Init() {
	if n.opts.id == "" {
		log.Fatal("instance id can not be empty")
	}
	if n.opts.codec == nil {
		log.Fatal("codec plugin is not injected")
	}
	if n.opts.locator == nil {
		log.Fatal("locator plugin is not injected")
	}
	if n.opts.registry == nil {
		log.Fatal("registry plugin is not injected")
	}
	if n.opts.transporter == nil {
		log.Fatal("transporter plugin is not injected")
	}
}

// Start 启动节点
func (n *Node) Start() {
	n.startRPCServer()

	n.registerInstance()

	n.proxy.watch(n.ctx)

	go n.dispatch()

	n.debugPrint()
}

// Destroy 销毁网关服务器
func (n *Node) Destroy() {
	n.deregisterInstance()

	n.stopRPCServer()

	close(n.chEvent)
	close(n.chRequest)
	n.cancel()
}

// Proxy 获取节点代理
func (n *Node) Proxy() Proxy {
	return n.proxy
}

// 分发处理消息
func (n *Node) dispatch() {
	for {
		select {
		case entity, ok := <-n.chEvent:
			if !ok {
				return
			}

			handler, ok := n.events[entity.event]
			if !ok {
				log.Warnf("event does not register handler function, event: %v", entity.event)
				continue
			}

			handler(entity.gid, entity.uid)
		case req, ok := <-n.chRequest:
			if !ok {
				return
			}

			n.rw.RLock()
			route, ok := n.routes[req.route]
			n.rw.RUnlock()

			if ok {
				route.handler(req)
			} else if n.defaultRouteHandler != nil {
				n.defaultRouteHandler(req)
			} else {
				log.Warnf("message routing does not register handler function, route: %v", req.route)
			}
		}
	}
}

// 启动RPC服务器
func (n *Node) startRPCServer() {
	var err error

	n.rpc, err = n.opts.transporter.NewNodeServer(&provider{n})
	if err != nil {
		log.Fatalf("the rpc server build failed: %v", err)
	}

	go func() {
		if err = n.rpc.Start(); err != nil {
			log.Fatalf("the rpc server startup failed: %v", err)
		}
	}()
}

// 停止RPC服务器
func (n *Node) stopRPCServer() {
	if err := n.rpc.Stop(); err != nil {
		log.Errorf("the rpc server stop failed: %v", err)
	}
}

// 注册服务实例
func (n *Node) registerInstance() {
	n.rw.RLock()
	routes := make([]registry.Route, 0, len(n.routes))
	for _, entity := range n.routes {
		routes = append(routes, registry.Route{
			ID:       entity.route,
			Stateful: entity.stateful,
		})
	}
	n.rw.RUnlock()

	n.instance = &registry.ServiceInstance{
		ID:       n.opts.id,
		Name:     string(cluster.Node),
		Kind:     cluster.Node,
		Alias:    n.opts.name,
		State:    cluster.Work,
		Routes:   routes,
		Endpoint: n.rpc.Endpoint().String(),
	}

	ctx, cancel := context.WithTimeout(n.ctx, 10*time.Second)
	err := n.opts.registry.Register(ctx, n.instance)
	cancel()
	if err != nil {
		log.Fatalf("the node service instance register failed: %v", err)
	}

	ctx, cancel = context.WithTimeout(n.ctx, 10*time.Second)
	watcher, err := n.opts.registry.Watch(ctx, string(cluster.Gate))
	cancel()
	if err != nil {
		log.Fatalf("the gate service watch failed: %v", err)
	}

	go func() {
		defer watcher.Stop()
		for {
			select {
			case <-n.ctx.Done():
				return
			default:
				// exec watch
			}
			services, err := watcher.Next()
			if err != nil {
				continue
			}
			n.router.ReplaceServices(services...)
		}
	}()
}

// 解注册服务实例
func (n *Node) deregisterInstance() {
	ctx, cancel := context.WithTimeout(n.ctx, 10*time.Second)
	err := n.opts.registry.Deregister(ctx, n.instance)
	cancel()
	if err != nil {
		log.Errorf("the node service instance deregister failed: %v", err)
	}
}

// 添加路由处理器
func (n *Node) addRouteHandler(route int32, stateful bool, handler RouteHandler) {
	n.rw.Lock()
	defer n.rw.Unlock()

	n.routes[route] = routeEntity{
		route:    route,
		stateful: stateful,
		handler:  handler,
	}
}

// 是否为有状态路由
func (n *Node) checkRouteStateful(route int32) (bool, bool) {
	n.rw.RLock()
	defer n.rw.RUnlock()

	if entity, ok := n.routes[route]; ok {
		return entity.stateful, ok
	}

	return false, n.defaultRouteHandler != nil
}

// 添加事件处理器
func (n *Node) addEventListener(event cluster.Event, handler EventHandler) {
	n.events[event] = handler
}

// 触发事件
func (n *Node) triggerEvent(event cluster.Event, gid string, uid int64) {
	n.chEvent <- &eventEntity{
		event: event,
		gid:   gid,
		uid:   uid,
	}
}

// 投递消息给当前节点处理
func (n *Node) deliverRequest(req *request) {
	n.chRequest <- req
}

func (n *Node) debugPrint() {
	log.Debugf("The node server startup successful")
	log.Debugf("RPC server, listen: %s protocol: %s", xnet.FulfillAddr(n.rpc.Addr()), n.rpc.Scheme())
}
