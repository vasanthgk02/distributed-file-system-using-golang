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
	log.Printf("new connection from %s (outbound: %v)\n", conn.RemoteAddr(), outbound)

	var err error

	defer func() {
		if err != nil {
			log.Printf("dropping peer connection %s: %v\n", conn.RemoteAddr(), err)
		} else {
			log.Printf("peer connection %s closed gracefully\n", conn.RemoteAddr())
		}
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.HandShakeFunc(peer); err != nil {
		log.Printf("handshake failed with %s: %v\n", conn.RemoteAddr(), err)
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer, outbound); err != nil {
			log.Printf("OnPeer callback failed for %s: %v\n", conn.RemoteAddr(), err)
			return
		}
	}

	// Message processing loop
	for {
		rpc := RPC{}
		rpc.From = conn.RemoteAddr() // Use RemoteAddr for better identification

		if err = t.Decoder.Decode(conn, &rpc); err != nil {
			log.Printf("decode error from %s: %v\n", conn.RemoteAddr(), err)
			return
		}

		log.Printf("decoded message from %s: stream=%v\n", conn.RemoteAddr(), rpc.Stream)

		if rpc.Stream {
			peer.wg.Add(1)
			log.Printf("[%s] incoming stream, waiting...\n", conn.RemoteAddr())
			peer.wg.Wait()
			log.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
			continue
		}
		t.rpcch <- rpc

	}

}
