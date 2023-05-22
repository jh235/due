package packet

import (
	"encoding/binary"
	"sync"
)

const (
	deliver int8 = iota + 1
)

type packer struct {
	byteOrder binary.ByteOrder
	buffers   sync.Pool
}

func () {

}

func (p *packer) Pack(id string, cid int64, uid int64, seq int32, route int32, buffer []byte) ([]byte, error) {

}
