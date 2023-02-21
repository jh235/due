package node_test

import (
	"github.com/dobyte/due/transport/rpcx/internal/server"
	"github.com/dobyte/due/transport/rpcx/node"
	"testing"
)

func TestNewServer(t *testing.T) {
	s, err := node.NewServer(nil, &server.Options{
		Addr: ":3554",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = s.Start(); err != nil {
		t.Fatal(err)
	}
}
