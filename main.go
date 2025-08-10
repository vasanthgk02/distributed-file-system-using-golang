package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/vasanthgk02/distributed_file_system/p2p"
)

func startServer1() {
	s1 := makeServer("127.0.0.1:5001")
	go func() {
		if err := s1.Start(); err != nil {
			panic(err)
		}
	}()
	select {}
}

func startServer2() {
	s2 := makeServer("127.0.0.1:5002", "127.0.0.1:5001")

	time.Sleep(1 * time.Second)
	go s2.Start()
	time.Sleep(5 * time.Second)

	file, _ := os.Open("/Users/vasanthgk02/Desktop/distribute_data.txt")
	fileInfo, _ := file.Stat()
	buff := make([]byte, fileInfo.Size())
	file.Read(buff)

	key := "vasanth"
	data := bytes.NewReader(buff)
	s2.Store(key, data)

	time.Sleep(time.Second * 5)

	// s2.Delete(key)
	// time.Sleep(time.Second * 5)

	d := make([]byte, 100)
	r, err := s2.Get("vasanth")

	if err != nil {
		panic(err)
	}
	r.Read(d)
	fmt.Println(string(d))
	select {}
}

func main() {
	// startServer1()
	startServer2()
}

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		HandShakeFunc: p2p.NOHandShake,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fileServerOpts := FileServerOpts{
		EncKey:            newEncryptionKey(),
		StorageRoot:       listenAddr + "_network",
		PathTransfromFunc: CASPathTransform,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}
