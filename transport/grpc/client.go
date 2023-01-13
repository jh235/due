package grpc

import (
	"context"
	"encoding/json"
	"github.com/dobyte/due/transport/grpc/internal/pb"
	"google.golang.org/protobuf/proto"
)

type client struct {
	client pb.ClusterClient
}

// Invoke 触发事件
func (c *client) Invoke(ctx context.Context, method string, args, reply interface{}) (err error) {
	req := &pb.InvokeRequest{Method: method}

	switch v := args.(type) {
	case proto.Message:
		req.Arguments, err = proto.Marshal(v)
		req.Encoding = pb.Encoding_Proto
	default:
		req.Arguments, err = json.Marshal(v)
		req.Encoding = pb.Encoding_Json
	}
	if err != nil {
		return
	}

	_, err = c.client.Invoke(ctx, req)
	if err != nil {
		return
	}

	return
}
