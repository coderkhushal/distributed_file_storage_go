package p2p

// message holds any arbitrary data that is being sent
// over each transport between 2 nodes in the network
type Message struct {
	Payload []byte
}
