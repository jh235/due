package netpoll

type ServerOption func(o *serverOptions)

type serverOptions struct {
	network string // 网络
	addr    string // 监听地址
}

func defaultServerOptions() *serverOptions {
	return &serverOptions{}
}
