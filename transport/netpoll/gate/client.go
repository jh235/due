package gate

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/dobyte/due/transport"
	"github.com/dobyte/due/utils/xconv"
	"sync"
)

const (
	deliver int8 = iota + 1
)

const (
	stateGID uint8 = 1 << iota
	stateCID
	stateUID
	stateSeq
)

type Client struct {
	byteOrder binary.ByteOrder
	buffers   sync.Pool
}

func NewClient() *Client {
	return &Client{
		byteOrder: binary.BigEndian,
		buffers: sync.Pool{New: func() interface{} {
			buf := &bytes.Buffer{}
			return buf
		}},
	}
}

// Deliver 投递消息
// protocol(int8) + state(int8)

func (c *Client) Deliver(ctx context.Context, args *transport.DeliverArgs) (miss bool, err error) {
	buf := c.buffers.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		c.buffers.Put(buf)
	}()

	var state uint8

	if args.GID != "" {
		state |= stateGID
	}

	if args.CID != 0 {
		state |= stateCID
	}

	if args.UID != 0 {
		state |= stateUID
	}

	if args.Message != nil && args.Message.Seq != 0 {
		state |= stateSeq
	}

	if err = binary.Write(buf, c.byteOrder, deliver); err != nil {
		return
	}

	if err = binary.Write(buf, c.byteOrder, state); err != nil {
		return
	}

	var clusterID []byte

	if args.GID != "" {
		clusterID = xconv.Bytes(args.GID)
	} else {
		clusterID = xconv.Bytes(args.NID)
	}

	if err = binary.Write(buf, c.byteOrder, int8(len(clusterID))); err != nil {
		return
	}

	if err = binary.Write(buf, c.byteOrder, clusterID); err != nil {
		return
	}

	fmt.Println(len([]byte(args.NID)))

	fmt.Println(state)

	//binary.Write(buf, c.byteOrder, args.GID)
	//
	//binary.Write(buf, c.byteOrder, args.GID)
	//
	//binary.Write(buf, c.byteOrder, args.CID)

	return
}
