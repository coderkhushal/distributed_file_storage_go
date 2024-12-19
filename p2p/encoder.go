package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(r io.Reader, msg *Message) error
}

type GOBDecoder struct {
}

func (d GOBDecoder) Decode(r io.Reader, msg *Message) error {
	return gob.NewDecoder(r).Decode(msg)
}

type NOPDecoder struct{}

func (d NOPDecoder) Decode(r io.Reader, msg *Message) error {
	buf := make([]byte, 1028)

	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	msg.Payload = buf[:n]
	return nil
}
