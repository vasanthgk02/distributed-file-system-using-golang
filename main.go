package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/vasanthgk02/distributed_file_system/p2p"
)

func startBootstrapNode(addr string) {
	log.Printf("Starting bootstrap node on %s\n", addr)
	s := makeServer(addr)
	if err := s.Start(); err != nil {
		panic(err)
	}
}

func startRegularNode(addr string, bootstrapNodes []string) {
	log.Printf("Starting node on %s, connecting to bootstrap nodes: %v\n", addr, bootstrapNodes)
	s := makeServer(addr, bootstrapNodes...)

	go func() {
		if err := s.Start(); err != nil {
			panic(err)
		}
	}()

	// Give server time to start and connect to bootstrap
	time.Sleep(2 * time.Second)
	log.Printf("Node started successfully. Connected to %d bootstrap node(s).\n", len(bootstrapNodes))
	fmt.Println("Available commands:")
	fmt.Println("  store <key> <file_path_or_data>  - Store file or data with key")
	fmt.Println("  get <key>                        - Retrieve data by key")
	fmt.Println("  delete <key>                     - Delete data by key")
	fmt.Println("  quit                             - Exit the program")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  store doc /home/user/document.txt")
	fmt.Println("  store msg \"Hello World\"")
	fmt.Println("  get doc")
	fmt.Println()

	// Start interactive CLI
	startCLI(s)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./fs --mode=<bootstrap|node> --port=<port> [--bootstrap=<addr1,addr2,...>]")
		fmt.Println("Examples:")
		fmt.Println("  ./fs --mode=bootstrap --port=:3001")
		fmt.Println("  ./fs --mode=bootstrap --port=:3002")
		fmt.Println("  ./fs --mode=node --port=3000 --bootstrap=:3001,:3002")
		os.Exit(1)
	}

	var mode, port, bootstrapStr string
	for _, arg := range os.Args[1:] {
		log.Println(arg)
		if len(arg) > 7 && arg[:7] == "--mode=" {
			mode = arg[7:]
		} else if len(arg) > 7 && arg[:7] == "--port=" {
			port = arg[7:]
		} else if len(arg) > 12 && arg[:12] == "--bootstrap=" {
			bootstrapStr = arg[12:]
		}
	}

	if mode == "" || port == "" {
		fmt.Println("Error: --mode and --port are required")
		os.Exit(1)
	}

	switch mode {
	case "bootstrap":
		startBootstrapNode(port)
	case "node":
		if bootstrapStr == "" {
			fmt.Println("Error: --bootstrap is required for node mode")
			os.Exit(1)
		}
		// Parse comma-separated bootstrap addresses
		bootstrapNodes := parseBootstrapNodes(bootstrapStr)
		startRegularNode(port, bootstrapNodes)
	default:
		fmt.Printf("Error: unknown mode '%s'. Use 'bootstrap' or 'node'\n", mode)
		os.Exit(1)
	}
}

func parseBootstrapNodes(bootstrapStr string) []string {
	if bootstrapStr == "" {
		return nil
	}
	// Split by comma and trim spaces
	nodes := make([]string, 0)
	for _, node := range strings.Split(bootstrapStr, ",") {
		node = strings.TrimSpace(node)
		if node != "" {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func startCLI(s *FileServer) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			fmt.Print("> ")
			continue
		}
		
		parts := strings.Fields(input)
		if len(parts) == 0 {
			fmt.Print("> ")
			continue
		}
		
		command := strings.ToLower(parts[0])
		
		switch command {
		case "store":
			if len(parts) < 3 {
				fmt.Println("Usage: store <key> <file_path_or_data>")
				fmt.Println("Examples:")
				fmt.Println("  store myfile /path/to/file.txt")
				fmt.Println("  store mydata \"Hello World\"")
			} else {
				key := parts[1]
				pathOrData := strings.Join(parts[2:], " ")
				handleStore(s, key, pathOrData)
			}
			
		case "get":
			if len(parts) < 2 {
				fmt.Println("Usage: get <key>")
			} else {
				key := parts[1]
				handleGet(s, key)
			}
			
		case "delete":
			if len(parts) < 2 {
				fmt.Println("Usage: delete <key>")
			} else {
				key := parts[1]
				handleDelete(s, key)
			}
			
		case "quit", "exit":
			fmt.Println("Goodbye!")
			s.Stop()
			os.Exit(0)
			
		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Available commands: store <key> <file_path_or_data>, get <key>, delete <key>, quit")
		}
		
		fmt.Print("> ")
	}
}

func handleStore(s *FileServer, key, pathOrData string) {
	var reader io.Reader
	var dataSize int64
	
	// Check if it's a file path
	if fileInfo, err := os.Stat(pathOrData); err == nil && !fileInfo.IsDir() {
		// It's a valid file path
		file, err := os.Open(pathOrData)
		if err != nil {
			fmt.Printf("Error opening file '%s': %v\n", pathOrData, err)
			return
		}
		defer file.Close()
		
		reader = file
		dataSize = fileInfo.Size()
		fmt.Printf("Storing file '%s' (%d bytes) with key '%s'...\n", pathOrData, dataSize, key)
	} else {
		// Treat as raw data
		reader = bytes.NewReader([]byte(pathOrData))
		dataSize = int64(len(pathOrData))
		fmt.Printf("Storing data (%d bytes) with key '%s'...\n", dataSize, key)
	}
	
	if err := s.Store(key, reader); err != nil {
		fmt.Printf("Error storing: %v\n", err)
	} else {
		fmt.Printf("Successfully stored '%s'\n", key)
	}
}

func handleGet(s *FileServer, key string) {
	reader, err := s.Get(key)
	if err != nil {
		fmt.Printf("Error getting file: %v\n", err)
		return
	}
	
	data, err := io.ReadAll(reader)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	
	fmt.Printf("Retrieved '%s': %s\n", key, string(data))
}

func handleDelete(s *FileServer, key string) {
	if err := s.Delete(key); err != nil {
		fmt.Printf("Error deleting file: %v\n", err)
	} else {
		fmt.Printf("Successfully deleted '%s'\n", key)
	}
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
