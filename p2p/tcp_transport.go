package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer represents a remote node over a tcp established connection
type TCPPeer struct {
	// underlying connection of the peer which in this case is
	// a tcp connection
	net.Conn
	Wg *sync.WaitGroup
	// if we dial and retrieve a connection => outbound == true
	// if we accept and retrieve a connection => outbound == false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		Wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC, 1024),
	}

}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// consume implements transport interface , which will return read-only channel
// for reading incoming messages recieved from other peer in the network
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// Dial implements the transport interface
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	return nil
}
func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	t.listener = ln

	go t.StartAcceptLoop()
	log.Printf("TCP transport listening on port %s\n", t.ListenAddr)
	return nil
}

func (t *TCPTransport) StartAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)

		}

		// accepting
		go t.handleConn(conn, false)
		fmt.Printf("new incoming connection %+v \n", conn)
	}
}

// Addr implements transport interface , return the address the
// transport is accepting conneections.
func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		fmt.Printf("dropping peer connection %s", err)
		conn.Close()
	}()
	peer := NewTCPPeer(conn, true)

	if err = t.HandshakeFunc(peer); err != nil {

		fmt.Printf("tcp handshake error %s \n", err)
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read loop
	for {
		rpc := RPC{}
		if err = t.Decoder.Decode(conn, &rpc); err != nil {
			fmt.Printf("TCP error %s \n", err)
			return
		}

		rpc.From = conn.RemoteAddr().String()
		if rpc.Stream {

			peer.Wg.Add(1)
			fmt.Printf("(%s) incoming stream waiting...\n", conn.RemoteAddr())
			peer.Wg.Wait()
			fmt.Printf("(%s) stream closed, resuming read loop...\n", conn.RemoteAddr())
			continue

		}
		t.rpcch <- rpc
	}
}
