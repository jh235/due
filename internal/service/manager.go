package service

import (
	"context"
	"github.com/dobyte/due/errors"
	"github.com/dobyte/due/log"
)

var (
	ErrNotFoundMethod   = errors.New("not found method")
	ErrNotFoundService  = errors.New("not found service")
	ErrNoCallableMethod = errors.New("no callable method")
)

type Manager struct {
	services map[string]*Service
}

func NewManager() *Manager {
	return &Manager{
		services: make(map[string]*Service),
	}
}

// Register 注册服务
func (m *Manager) Register(providers ...interface{}) {
	for _, provider := range providers {
		service, err := ParseServiceProvider(provider)
		if err != nil {
			log.Warnf("register service failed: %v", err)
			continue
		}
		m.services[service.name] = service
	}
}

// Call 调用服务
func (m *Manager) Call(ctx context.Context, service, method string, args interface{}) (interface{}, error) {
	svc, ok := m.services[service]
	if !ok {
		return nil, ErrNotFoundService
	}

	return svc.Call(ctx, method, args)
}
