package p2p

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// RPC holds any arbitrary data that is being sent
// over each transport between 2 nodes in the network
type RPC struct {
	Payload []byte
	From    string
	Stream  bool
}
