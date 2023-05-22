package netpoll

import (
	"context"
	"github.com/cloudwego/netpoll"
	"github.com/dobyte/due/network"
	"time"
)

type server struct {
	opts              *serverOptions            // 配置
	listener          netpoll.Listener          // 监听器
	eventLoop         netpoll.EventLoop         // 事件驱动调度器
	startHandler      network.StartHandler      // 服务器启动hook函数
	stopHandler       network.CloseHandler      // 服务器关闭hook函数
	connectHandler    network.ConnectHandler    // 连接打开hook函数
	disconnectHandler network.DisconnectHandler // 连接关闭hook函数
	receiveHandler    network.ReceiveHandler    // 接收消息hook函数
}

// Start 启动服务器
func (s *server) Start() error {
	ln, err := netpoll.CreateListener(s.opts.network, s.opts.addr)
	if err != nil {
		return err
	}

	eventLoop, err := netpoll.NewEventLoop(
		func(ctx context.Context, conn netpoll.Connection) error {

		},
		netpoll.WithOnConnect(func(ctx context.Context, conn netpoll.Connection) context.Context {
			if s.connectHandler != nil {
				s.connectHandler()
			}

			return ctx
		}),
		netpoll.WithReadTimeout(time.Second),
	)
	if err != nil {
		return err
	}

	s.listener = ln
	s.eventLoop = eventLoop

	return eventLoop.Serve(s.listener)
}

// Stop 关闭服务器
func (s *server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.eventLoop.Shutdown(ctx)
}

// OnStart 监听服务器启动
func (s *server) OnStart(handler network.StartHandler) {
	s.startHandler = handler
}

// OnStop 监听服务器关闭
func (s *server) OnStop(handler network.CloseHandler) {
	s.stopHandler = handler
}

// OnConnect 监听连接打开
func (s *server) OnConnect(handler network.ConnectHandler) {
	s.connectHandler = handler
}

// OnDisconnect 监听连接关闭
func (s *server) OnDisconnect(handler network.DisconnectHandler) {
	s.disconnectHandler = handler
}

// OnReceive 监听接收到消息
func (s *server) OnReceive(handler network.ReceiveHandler) {
	s.receiveHandler = handler
}
