package p2p

import (
	"log"
	"net"
	"sync"
)

type TCPPeer struct {
	// The underlying connection of the peer which in this case
	// is the TCP connection
	net.Conn

	// if we dail and retrive a conn => outbound == true
	// if we accept and retrive a conn => outbound == false
	outbound bool

	wg *sync.WaitGroup
}

func (peer *TCPPeer) CloseStream() {
	peer.wg.Done()
}

func (peer *TCPPeer) Send(b []byte) error {
	_, err := peer.Conn.Write(b)
	return err
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{Conn: conn, outbound: outbound, wg: &sync.WaitGroup{}}
}

type TCPTransportOpts struct {
	ListenAddress string
	HandShakeFunc HandShakeFunc
	Decoder       Decoder
	OnPeer        func(Peer, bool) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC

	// mu    sync.RWMutex
	// peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC, 1024),
	}
}

func (t *TCPTransport) ListenAddr() string {
	return t.ListenAddress
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}
	go t.startAcceptLoop()
	log.Printf("TCP connection is listening on port: %s\n", t.ListenAddress)

	return nil

}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			log.Printf("TCP accept error: %s\n", err)
			return
		}
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) Dail(addr string) error {

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	log.Println("new conn triggered!")

	var err error

	defer func() {
		log.Printf("dropping peer connection: %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.HandShakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer, outbound); err != nil {
			return
		}
	} else {
		log.Println("onPeer is nil")
	}

	for {
		rpc := RPC{}
		rpc.From = conn.LocalAddr()
		if err = t.Decoder.Decode(conn, &rpc); err != nil {
			// log.Println(reflect.TypeOf(err))
			// panic(err)
			log.Printf("TCP error: %s\n", err)
			return
		}

		log.Printf("Decoded msg from conn. Sending same msg into chann: %+v\n", rpc)
		if rpc.Stream {

			peer.wg.Add(1)
			log.Printf("[%s] incoming stream, waiting...\n", conn.LocalAddr())
			peer.wg.Wait()
			log.Printf("[%s] stream closed, resuming read loop\n", conn.LocalAddr())
			continue
		}
		t.rpcch <- rpc

	}

}
