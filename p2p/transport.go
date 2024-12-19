package p2p

// peer is a interface that represent the remote node
type Peer interface {
}

// transport is anything which handles the communication
// between the nodes in the network. This can be of the
// form (TCP, UDP, Websockets , ....)
type Transport interface {
}
