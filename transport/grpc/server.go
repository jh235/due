package grpc

import (
	"context"
	endpointpkg "github.com/dobyte/due/internal/endpoint"
	"github.com/dobyte/due/internal/service"
	"github.com/dobyte/due/transport/grpc/internal/pb"
	"github.com/dobyte/due/utils/xnet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"net"
	"strings"
)

type Server struct {
	addr       string
	endpoint   *endpointpkg.Endpoint
	lis        net.Listener
	server     *grpc.Server
	serviceMgr *service.Manager
}

type Options struct {
	Addr       string
	KeyFile    string
	CertFile   string
	ServerOpts []grpc.ServerOption
}

func NewServer(opts *Options) (*Server, error) {
	host, port, err := net.SplitHostPort(opts.Addr)
	if err != nil {
		return nil, err
	}

	var (
		addr       string
		isSecure   = false
		serverOpts = make([]grpc.ServerOption, 0, len(opts.ServerOpts)+1)
		server     = &Server{}
	)

	if len(host) > 0 && (host != "0.0.0.0" && host != "[::]" && host != "::") {
		server.addr = net.JoinHostPort(host, port)
		addr = server.addr
	} else {
		server.addr = net.JoinHostPort("", port)
		if ip, err := xnet.InternalIP(); err != nil {
			return nil, err
		} else {
			addr = net.JoinHostPort(ip, port)
		}
	}

	serverOpts = append(serverOpts, opts.ServerOpts...)
	if opts.CertFile != "" && opts.KeyFile != "" {
		cred, err := credentials.NewServerTLSFromFile(opts.CertFile, opts.KeyFile)
		if err != nil {
			return nil, err
		}
		serverOpts = append(serverOpts, grpc.Creds(cred))
		isSecure = true
	}

	server.server = grpc.NewServer(serverOpts...)
	server.endpoint = endpointpkg.NewEndpoint("grpc", addr, isSecure)
	server.serviceMgr = service.NewManager()
	server.server.RegisterService(&pb.Cluster_ServiceDesc, &endpoint{serviceMgr: server.serviceMgr})

	return server, nil
}

// Addr 监听地址
func (s *Server) Addr() string {
	return s.addr
}

// Scheme 协议
func (s *Server) Scheme() string {
	return s.endpoint.Scheme()
}

// Endpoint 获取服务端口
func (s *Server) Endpoint() *endpointpkg.Endpoint {
	return s.endpoint
}

// Start 启动服务器
func (s *Server) Start() error {
	addr, err := net.ResolveTCPAddr("tcp", s.addr)
	if err != nil {
		return err
	}

	s.lis, err = net.Listen(addr.Network(), addr.String())
	if err != nil {
		return err
	}

	return s.server.Serve(s.lis)
}

// Stop 停止服务器
func (s *Server) Stop() error {
	s.server.Stop()
	return s.lis.Close()
}

// RegisterServiceProvider 注册服务
func (s *Server) RegisterServiceProvider(providers ...interface{}) {
	s.serviceMgr.Register(providers...)
}

type endpoint struct {
	pb.UnimplementedClusterServer
	serviceMgr *service.Manager
}

// Invoke 触发事件
func (e *endpoint) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeReply, error) {
	paths := strings.SplitN(req.Path, ".", 2)
	if len(paths) != 2 {
		return nil, status.New(codes.InvalidArgument, "invalid argument").Err()
	}

	switch req.Encoding {
	case pb.Encoding_Proto:
	}

	e.serviceMgr.Call(ctx, paths[0], paths[1])
}
