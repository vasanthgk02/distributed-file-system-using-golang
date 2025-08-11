package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/vasanthgk02/distributed_file_system/p2p"
)

type FILE_ACTION string

const (
	ACTION_DELETE FILE_ACTION = "DELETE"
	ACTION_GET    FILE_ACTION = "GET"
)

const (
	FILE_NOT_FOUND int64 = -1
)

type FileServerOpts struct {
	EncKey         []byte
	Transport      p2p.Transport
	BootstrapNodes []string

	// StoreOpts
	StorageRoot       string
	PathTransfromFunc PathTransfromFunc
}

type FileServer struct {
	FileServerOpts

	store  *Store
	quitch chan struct{}

	peerLock sync.Mutex
	peers    map[string]p2p.Peer
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransfromFunc: opts.PathTransfromFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) broadCast(msg *Message) error {
	log.Printf("broadcasting msg: %+v", *msg)
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send([]byte{p2p.IncomingMessage}); err == nil {
			if err := peer.Send(buf.Bytes()); err != nil {
				log.Printf("Error sending message to peer [%s]", peer.LocalAddr())
			}
		} else {
			log.Printf("error: unable brodcast msg to: [%s]", peer.LocalAddr())
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

type MessageFileKey struct {
	Key    string
	Action FILE_ACTION
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		log.Printf("[%s] serving file [%s] from local disk\n", s.Transport.ListenAddr(), key)
		_, r, err := s.store.Read(key)
		if err != nil {
			return nil, err
		}
		return r, err
	}
	log.Printf("[%s] dont have [%s] file locally, fetching from network", s.Transport.ListenAddr(), key)

	msg := Message{
		Payload: MessageFileKey{
			Key:    hashKey(key),
			Action: ACTION_GET,
		},
	}

	// peerWriters := s.peersWriter()
	// mw := io.MultiWriter(peerWriters...)
	// mw.Write([]byte{p2p.IncomingMessage})

	if err := s.broadCast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 5)

	time.Sleep(time.Millisecond * 500)

	var filesize int64
	for _, peer := range s.peers {

		binary.Read(peer, binary.LittleEndian, &filesize)

		if filesize == -1 {
			log.Printf("file not found on server [%s]\n", peer.LocalAddr())
			continue
		}

		n, err := s.store.writeDecrypt(s.EncKey, key, io.LimitReader(peer, filesize))
		if err != nil {
			log.Printf("Error: [%s] failed while writing from peer: %s", s.Transport.ListenAddr(), peer.LocalAddr())
			continue
		}
		log.Printf("[%s] received bytes over the network %d from [%s]\n", s.Transport.ListenAddr(), n, peer.LocalAddr())
		peer.CloseStream()
	}

	if filesize == -1 {
		return nil, errors.New("key does not exist in network")
	}

	_, r, err := s.store.Read(key)
	return r, err
}

func (s *FileServer) Start() error {
	log.Printf("[%s] Starting file server...", s.Transport.ListenAddr())
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.bootstrapNetwork()
	s.loop()
	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to error or user quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			log.Printf("Receiver triggered. Current msg: %+v\n", rpc)
			log.Printf("payload string: %s\n", rpc.Payload)
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error:", err)
				continue
			}
			log.Printf("Received msg: %+v\n", msg)
			log.Println(reflect.TypeOf(msg.Payload))

			if err := s.handleMessage(rpc.From.String(), &msg); err != nil {
				log.Println("handle message error:", err)
			}
		case <-s.quitch:
			log.Println("quit msg received")
			return
		}
	}
}

func (s *FileServer) OnPeer(p p2p.Peer, outbound bool) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("current local peer: %+v, current remote peer: %+v\n", p.LocalAddr(), p.RemoteAddr())
	// log.Printf("connected with remote: %s\n", p.LocalAddr().String())

	return nil
}

func (s *FileServer) peersWriter() (ioWriter []io.Writer) {
	for _, peer := range s.peers {
		ioWriter = append(ioWriter, peer)
	}
	return ioWriter
}

func (s *FileServer) Store(key string, r io.Reader) error {

	var (
		fileBuff = new(bytes.Buffer)
		tee      = io.TeeReader(r, fileBuff)
	)

	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  hashKey(key),
			Size: size + 16,
		},
	}

	if err := s.broadCast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Second * 5)

	peersWriter := s.peersWriter()
	mw := io.MultiWriter(peersWriter...)
	mw.Write([]byte{p2p.IncomingStream})

	n, err := copyEncrypt(s.EncKey, fileBuff, mw)
	if err != nil {
		log.Println(err)
	}

	log.Printf("[%s] received and written (%d) bytes to disk\n", s.Transport.ListenAddr(), n)

	return nil

}

func (s *FileServer) Delete(key string) error {
	if s.store.Has(key) {
		s.store.Delete(key)
		log.Printf("file [%s] deleted from local\ndd", key)
	}
	log.Println("broadcasting msg over network to delete file from network")

	var msg Message = Message{
		Payload: MessageFileKey{
			Key:    hashKey(key),
			Action: ACTION_DELETE,
		},
	}

	peerWrites := s.peersWriter()
	mw := io.MultiWriter(peerWrites...)

	// Send message type
	mw.Write([]byte{p2p.IncomingMessage})

	// encode message
	buff := new(bytes.Buffer)
	if err := gob.NewEncoder(buff).Encode(msg); err != nil {
		log.Printf("error while encoding delete message\n%s\n", err)
		return err
	}

	_, err := mw.Write(buff.Bytes())
	if err != nil {
		log.Printf("error occured while deleting file [%s] from network\n%s\n", key, err)
		return err
	}
	return nil

}

func (s *FileServer) bootstrapNetwork() error {
	if len(s.BootstrapNodes) == 0 {
		return nil
	}
	for _, addr := range s.BootstrapNodes {
		log.Println("Running for addr:", addr)
		go func(addr string) {
			log.Printf("[%s] attemping to connect with remote %s\n", s.Transport.ListenAddr(), addr)
			if err := s.Transport.Dail(addr); err != nil {
				log.Printf("Error: dail err:\n%s\n", err)
			}
		}(addr)
	}
	return nil
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageFileKey:
		return s.handleMessageFileKey(from, v)
	default:
		log.Printf("message type not supported...\n")
	}

	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		log.Printf("peer map: %+v", s.peers)
		return fmt.Errorf("missing peer from peer map: %s", from)
	}
	defer peer.CloseStream()

	s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	return nil

}

func (s *FileServer) handleMessageFileKey(from string, msg MessageFileKey) error {

	switch msg.Action {

	case "GET":
		peer, ok := s.peers[from]
		if !ok {
			return fmt.Errorf("peer [%s] does not exist in peer map", from)
		}
		if !s.store.Has(msg.Key) {
			peer.Send([]byte{p2p.IncomingStream})
			binary.Write(peer, binary.LittleEndian, FILE_NOT_FOUND)
			return fmt.Errorf("[%s] need to serve file (%s) but it does not exist on disk", s.Transport.ListenAddr(), msg.Key)
		}

		log.Printf("[%s] serving file (%s) over the network\n", s.Transport.ListenAddr(), msg.Key)

		fileSize, r, err := s.store.Read(msg.Key)
		if err != nil {
			return err
		}

		if rc, ok := r.(io.ReadCloser); ok {
			log.Println("closing file...")
			defer rc.Close()
		}

		peer.Send([]byte{p2p.IncomingStream})
		binary.Write(peer, binary.LittleEndian, fileSize)
		n, err := io.Copy(peer, r)
		if err != nil {
			return err
		}
		log.Printf("[%s] written (%d) bytes over the network to %s\n", s.Transport.ListenAddr(), n, from)

		return nil

	case "DELETE":
		if !s.store.Has(msg.Key) {
			return fmt.Errorf("[%s] need to delete file (%s) but it does not exist on disk", s.Transport.ListenAddr(), msg.Key)
		}

		err := s.store.Delete(msg.Key)
		if err != nil {
			log.Printf("error occured while deleting file [%s]", msg.Key)
			return err
		}

		peer, ok := s.peers[from]
		if !ok {
			return fmt.Errorf("peer [%s] does not exist in peer map", from)
		}

		peer.Send([]byte{p2p.IncomingMessage})
		var ackSig Message = Message{
			Payload: fmt.Appendf(nil, "ACK DEL from [%s]", from),
		}
		buff := new(bytes.Buffer)
		if err := gob.NewEncoder(buff).Encode(ackSig); err != nil {
			return err
		}
		if err := peer.Send(buff.Bytes()); err != nil {
			return err
		}
		return nil
	default:
		log.Printf("unsupported action: [%s]\n", msg.Action)
	}
	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageFileKey{})
}
