package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	tcpOpts := TCPTransportOpts{
		ListenAddress: ":3000",
		HandShakeFunc: NOHandShake,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransport(tcpOpts)
	assert.Equal(t, tr.ListenAddress, ":3000")

	// Server
	assert.Nil(t, tr.ListenAndAccept())
}
