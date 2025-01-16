package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(r io.Reader, msg *RPC) error
}

type GOBDecoder struct {
}

func (d GOBDecoder) Decode(r io.Reader, msg *RPC) error {
	return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (d DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	peekbuf := make([]byte, 1)
	if _, err := r.Read(peekbuf); err != nil {
		return err
	}

	// In case of stream , we are not decoding what is sent over the network
	// we are just setting stream true so we can handle that in our logic
	stream := peekbuf[0] == IncomingStream

	if stream {
		msg.Stream = true
		return nil
	}

	buf := make([]byte, 1028)

	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	msg.Payload = buf[:n]
	return nil
}
