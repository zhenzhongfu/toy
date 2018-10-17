package network

import (
	"encoding/binary"
	"fmt"
)

const (
	HeaderLen = 12
)

type Message struct {
	Ctrl byte // TODO 控制位，'m':message, 'c':close
	Cmd  uint32
	Seq  uint32
	Size uint32
	Body []byte
}

func MakePack(cmd uint32, body []byte) []byte {
	size := len(body)
	b := make([]byte, size+16)
	binary.BigEndian.PutUint32(b[0:], uint32(size+12))
	binary.BigEndian.PutUint32(b[4:], uint32(cmd))
	binary.BigEndian.PutUint32(b[8:], uint32(0))
	binary.BigEndian.PutUint32(b[12:], uint32(size))
	copy(b[16:], body)
	return b
}

func (msg *Message) Pack() ([]byte, error) {
	b := make([]byte, msg.Size+16)
	binary.BigEndian.PutUint32(b[0:], uint32(msg.Size+12))
	binary.BigEndian.PutUint32(b[4:], uint32(msg.Cmd))
	binary.BigEndian.PutUint32(b[8:], uint32(msg.Seq))
	binary.BigEndian.PutUint32(b[12:], uint32(msg.Size))
	copy(b[16:], msg.Body)

	return b, nil
}

func (msg *Message) Unpack(b []byte) error {
	if len(b) < HeaderLen {
		return fmt.Errorf("unpack error. not enough len(%d)", len(b))
	}

	cmd := binary.BigEndian.Uint32(b[0:])
	seq := binary.BigEndian.Uint32(b[4:])
	size := binary.BigEndian.Uint32(b[8:])
	body := b[12:]

	msg.Ctrl = 'm'
	msg.Cmd = cmd
	msg.Seq = seq
	msg.Size = uint32(size)
	msg.Body = make([]byte, size)
	copy(msg.Body, body)
	return nil
}
