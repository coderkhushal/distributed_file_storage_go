package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/coderkhushal/alltimestore/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNode     []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeopts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeopts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) stream(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)

	}

	mw := io.MultiWriter(peers...)

	err := gob.NewEncoder(mw).Encode(msg)
	if err != nil {
		fmt.Println("error encoding payload, ", err)
		return err
	}

	return nil

}
func (s *FileServer) broadcast(msg *Message) error {

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err

	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

type MessageGetFile struct {
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		return s.store.Read(key)
	}
	fmt.Printf("dont have file %s locally , fetching from network...\n", key)

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 3)
	for _, peer := range s.peers {
		filebuffer := new(bytes.Buffer)
		n, err := io.CopyN(peer, filebuffer, 22)

		if err != nil {
			return nil, err

		}

		fmt.Printf("recieved %d bytes over network\n", n)
		fmt.Print(filebuffer.String())
	}
	select {}
	return nil, nil
}

func (s *FileServer) Store(key string, r io.Reader) error {

	filebuffer := new(bytes.Buffer)
	tee := io.TeeReader(r, filebuffer)
	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return err
	}
	time.Sleep(time.Second * 3)

	// TODOO : use a multiwriter here
	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingStream})
		n, err := io.Copy(peer, filebuffer)
		if err != nil {
			return err
		}
		fmt.Println("recieved and written to disk ", n)
	}
	return nil
}
func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("connected with remote %s \n", p)
	return nil
}

func (s *FileServer) loop() {
	defer func() {
		fmt.Println("file server stopped due to error or user quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():

			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error ", err)
			}

			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("handle messsage error ", err)
			}
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}

	return nil
}
func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {

	if !s.store.Has(msg.Key) {
		return fmt.Errorf("server required file (%s) but it does not exist in disk \n", msg.Key)

	}
	r, err := s.store.Read(msg.Key)

	if err != nil {
		return err
	}
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("Peer %s is not in map\n", from)

	}
	_, err = io.Copy(peer, r)
	if err != nil {
		return err

	}

	return nil
}
func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	fmt.Printf("recv store file msg : %+v \n", msg)
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list ", from)

	}
	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {

		return err
	}
	log.Printf("written (%d) bytes to disk  %s ", n, msg.Key)

	peer.(*p2p.TCPPeer).Wg.Done()
	return nil
}
func (s *FileServer) BootstrapNetwork() {

	for _, addr := range s.BootstrapNode {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			fmt.Println("attempting to connect with remote : ", addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("Dial error : ", err)
			}
		}(addr)
	}
}
func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	if len(s.BootstrapNode) != 0 {
		s.BootstrapNetwork()
	}

	s.loop()
	return nil
}

// func (s *FileServer) broadcast(p *Payload) error {
// }
