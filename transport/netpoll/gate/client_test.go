package gate_test

import (
	"context"
	"github.com/dobyte/due/transport"
	"github.com/dobyte/due/transport/netpoll/gate"
	"github.com/dobyte/due/utils/xuuid"
	"testing"
)

func TestClient_Deliver(t *testing.T) {
	c := gate.NewClient()

	uuid, _ := xuuid.UUID()

	c.Deliver(context.Background(), &transport.DeliverArgs{
		NID: uuid,
		CID: 1,
	})
}
