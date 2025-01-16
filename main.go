package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/coderkhushal/alltimestore/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {

	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: CACPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNode:     nodes,
	}
	s := NewFileServer(fileServerOpts)

	tcpTransport.OnPeer = s.OnPeer
	return s
}

func OnPeer(p2p.Peer) error {
	fmt.Println("doing some logic with peer outside of TCPTransport")
	return nil
}

func main() {
	s1 := makeServer(":3000")
	s2 := makeServer(":4000", ":3000")
	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(3 * time.Second)

	go s2.Start()
	time.Sleep(3 * time.Second)

	// data := bytes.NewReader([]byte("my big data files here"))
	// s2.Store("myprivatedata", data)
	r, err := s2.Get("myprivatedata")
	if err != nil {
		log.Fatal(err)

	}

	b, err := io.ReadAll(r)
	if err != nil {

		log.Fatal(err)
	}
	fmt.Println(string(b))
	select {}
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
