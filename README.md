# Distributed File System

A peer-to-peer distributed file system built in Go that enables secure file storage and retrieval across multiple nodes in a network.

## Features

- **P2P Network Architecture**: Decentralized file storage across multiple nodes
- **Content-Addressable Storage (CAS)**: Files are stored using SHA-1 hash-based paths
- **AES Encryption**: All files are encrypted before storage and transmission
- **TCP Transport Layer**: Reliable communication between nodes
- **Automatic File Replication**: Files are automatically replicated across network peers
- **Network File Discovery**: Automatic file retrieval from remote nodes when not available locally

## Architecture

### Core Components

- **FileServer**: Main server handling file operations and peer management
- **P2P Transport**: TCP-based peer-to-peer communication layer
- **Storage Engine**: Content-addressable storage with encryption support
- **Crypto Module**: AES encryption/decryption for secure file handling

### Network Protocol

The system uses a custom message-based protocol over TCP:
- `IncomingMessage`: Regular message communication
- `IncomingStream`: File data streaming
- Message types: `STORE`, `GET`, `DELETE`

## Installation

```bash
# Clone the repository
git clone https://github.com/vasanthgk02/distributed_file_system.git
cd distributed_file_system

# Build the project
make build

# Run the application
make run
```

## Usage

### Starting a Network

```go
// Start first node (bootstrap node)
s1 := makeServer("127.0.0.1:5001")
go s1.Start()

// Start second node and connect to bootstrap
s2 := makeServer("127.0.0.1:5002", "127.0.0.1:5001")
go s2.Start()
```

### File Operations

```go
// Store a file
key := "myfile"
data := bytes.NewReader(fileData)
err := server.Store(key, data)

// Retrieve a file
reader, err := server.Get("myfile")

// Delete a file
err := server.Delete("myfile")
```

## Configuration

### Server Options

```go
type FileServerOpts struct {
    EncKey            []byte              // Encryption key for file security
    StorageRoot       string              // Local storage directory
    PathTransformFunc PathTransformFunc   // Path transformation function
    Transport         p2p.Transport       // Network transport layer
    BootstrapNodes    []string           // Initial nodes to connect to
}
```

### Transport Options

```go
type TCPTransportOpts struct {
    ListenAddress string              // Address to listen on
    HandShakeFunc HandShakeFunc       // Peer handshake function
    Decoder       Decoder             // Message decoder
    OnPeer        func(Peer, bool) error // Peer connection callback
}
```

## File Storage

Files are stored using Content-Addressable Storage (CAS):
- File keys are hashed using SHA-1
- Hash is split into directory structure (e.g., `abcde/fghij/klmno/...`)
- Files are encrypted with AES before storage
- Each node maintains its own storage directory

## Network Communication

### Message Types

1. **MessageStoreFile**: Notifies peers about new file storage
2. **MessageFileKey**: Requests file operations (GET/DELETE)

### File Transfer Protocol

1. Node broadcasts file availability
2. Peers request file if needed
3. File is streamed with size header
4. Encryption/decryption handled transparently

## Testing

```bash
# Run all tests
make test

# Run specific package tests
go test ./p2p -v
go test ./storage -v
```

## Project Structure

```
distributed_file_system/
├── bin/                    # Compiled binaries
├── p2p/                    # P2P networking layer
│   ├── transport.go        # Transport interface
│   ├── tcp_transport.go    # TCP implementation
│   ├── encoding.go         # Message encoding
│   ├── handshake.go        # Peer handshake
│   └── message.go          # Message types
├── main.go                 # Application entry point
├── server.go               # File server implementation
├── storage.go              # Storage engine
├── crypto.go               # Encryption utilities
├── Makefile               # Build configuration
└── go.mod                 # Go module definition
```

## Security Features

- **AES Encryption**: All files encrypted with 256-bit keys
- **Secure Key Generation**: Cryptographically secure random keys
- **Hash-based Addressing**: Content integrity through SHA-1 hashing
- **Peer Authentication**: Handshake protocol for peer verification
