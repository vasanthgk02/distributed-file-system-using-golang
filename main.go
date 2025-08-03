package main

import (
	"fmt"
	"log"

	"github.com/vasanthgk02/distributed_file_system/p2p"
)

func OnPeer(p p2p.Peer) error {
	p.Close()
	return nil
	// return fmt.Errorf("failed the onpeer func")
}

func main() {

	tcpOpts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		HandShakeFunc: p2p.NOHandShake,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	fmt.Printf("Starting TCP with opts: %v\n", tcpOpts)

	tr := p2p.NewTCPTransport(tcpOpts)

	err := tr.ListenAndAccept()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()
	select {}

}
