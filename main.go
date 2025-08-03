package main

import (
	"fmt"
	"log"

	"github.com/vasanthgk02/distributed_file_system/p2p"
)

func main() {

	tcpOpts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		HandShakeFunc: p2p.NOHandShake,
		Decoder:       p2p.DefaultDecoder{},
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	err := tr.ListenAndAccept()
	if err != nil {
		log.Fatal(err)
	}
	select {}

	fmt.Print("Distributed File System")
}
