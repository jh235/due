package service

import (
	"context"
	"github.com/dobyte/due/errors"
	"github.com/dobyte/due/log"
	"github.com/dobyte/due/utils/xstring"
	"reflect"
	"runtime"
)

var (
	errorReflectType   = reflect.TypeOf((*error)(nil)).Elem()
	contextReflectType = reflect.TypeOf((*context.Context)(nil)).Elem()
)

type Service struct {
	typ     reflect.Type
	rcv     reflect.Value
	name    string
	methods map[string]*Method
}

type Method struct {
	method reflect.Method
	args   reflect.Type
}

// Call 调用方法
func (s *Service) Call(ctx context.Context, method string, args interface{}) (interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case runtime.Error:
				log.Panicf("runtime error: %v", err)
			default:
				log.Panic(err)
			}
		}
	}()

	m, ok := s.methods[method]
	if !ok {
		return nil, ErrNotFoundMethod
	}

	var values []reflect.Value

	if m.args.Kind() != reflect.Ptr {
		values = m.method.Func.Call([]reflect.Value{s.rcv, reflect.ValueOf(ctx), reflect.ValueOf(args).Elem()})
	} else {
		values = m.method.Func.Call([]reflect.Value{s.rcv, reflect.ValueOf(ctx), reflect.ValueOf(args)})
	}

	if err := values[1].Interface(); err != nil {
		return values[0].Interface(), err.(error)
	} else {
		return values[0].Interface(), nil
	}
}

// ParseServiceProvider 解析服务提供者
func ParseServiceProvider(provider interface{}) (*Service, error) {
	typ := reflect.TypeOf(provider)
	rcv := reflect.ValueOf(provider)
	num := typ.NumMethod()
	name := reflect.Indirect(rcv).Type().Name()

	if !xstring.FirstIsUpper(name) {
		return nil, errors.New("type " + name + " is not exported")
	}

	if num == 0 {
		return nil, errors.New("type " + name + " has no callable methods")
	}

	methods := make(map[string]*Method, num)
	for i := 0; i < num; i++ {
		method := typ.Method(i)

		if !method.IsExported() {
			continue
		}

		if method.Type.NumIn() != 3 {
			continue
		}

		if method.Type.NumOut() != 2 {
			continue
		}

		if !method.Type.In(1).Implements(contextReflectType) {
			continue
		}

		if !method.Type.Out(1).Implements(errorReflectType) {
			continue
		}

		methods[method.Name] = &Method{
			method: method,
			args:   method.Type.In(2),
		}
	}

	if len(methods) == 0 {
		return nil, errors.New("type " + name + " has no callable methods")
	}

	if i, ok := provider.(interface{ ServiceName() string }); ok {
		name = i.ServiceName()
	}

	return &Service{
		typ:     typ,
		rcv:     rcv,
		name:    name,
		methods: methods,
	}, nil
}
