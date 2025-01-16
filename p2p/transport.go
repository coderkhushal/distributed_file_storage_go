package p2p

import "net"

// peer is a interface that represent the remote node
type Peer interface {
	net.Conn

	Send([]byte) error
	RemoteAddr() net.Addr
}

// transport is anything which handles the communication
// between the nodes in the network. This can be of the
// form (TCP, UDP, Websockets , ....)
type Transport interface {
	Addr() string
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
	Dial(string) error
}
