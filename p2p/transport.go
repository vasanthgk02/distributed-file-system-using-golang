package p2p

import (
	"net"
)

type Peer interface {
	// Conn() net.Conn
	net.Conn
	Send([]byte) error
	CloseStream()
}

type Transport interface {
	Dail(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
	ListenAddr() string
}
