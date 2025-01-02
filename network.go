package main

import (
    "encoding/json"
    "fmt"
    "net"
    "sync"
)
// ChunkRequest represents a request for a specific chunk
type ChunkRequest struct {
	Hash string `json:"hash"`
}

// ChunkResponse represents a chunk response
type ChunkResponse struct {
	Hash string `json:"hash"`
	Data []byte `json:"data"`
}
type P2PNode struct {
    storage    *StorageEngine
    connMgr    *ConnectionManager
    listenAddr string
    listener   net.Listener
    wg         sync.WaitGroup
    stopping   bool
    stopMutex  sync.RWMutex
}

func NewP2PNode(listenAddr string) (*P2PNode, error) {
    storage, err := NewStorageEngine()
    if err != nil {
        return nil, fmt.Errorf("failed to create storage engine: %v", err)
    }

    return &P2PNode{
        storage:    storage,
        connMgr:    NewConnectionManager(),
        listenAddr: listenAddr,
    }, nil
}

func (n *P2PNode) Start() error {
    listener, err := net.Listen("tcp", n.listenAddr)
    if err != nil {
        return fmt.Errorf("failed to start listener: %v", err)
    }
    n.listener = listener
    
    // Update listen address with actual port
    n.listenAddr = listener.Addr().String()
    
    fmt.Printf("P2P node listening on %s\n", n.listenAddr)

    n.wg.Add(1)
    go n.acceptConnections()

    return nil
}

func (n *P2PNode) acceptConnections() {
    defer n.wg.Done()
    
    for {
        n.stopMutex.RLock()
        if n.stopping {
            n.stopMutex.RUnlock()
            return
        }
        n.stopMutex.RUnlock()

        conn, err := n.listener.Accept()
        if err != nil {
            if !n.isStopping() {
                fmt.Printf("Failed to accept connection: %v\n", err)
            }
            continue
        }
        
        go n.handleConnection(conn)
    }
}

func (n *P2PNode) isStopping() bool {
    n.stopMutex.RLock()
    defer n.stopMutex.RUnlock()
    return n.stopping
}

func (n *P2PNode) Stop() {
    n.stopMutex.Lock()
    n.stopping = true
    n.stopMutex.Unlock()

    if n.listener != nil {
        n.listener.Close()
    }
    n.wg.Wait()
}

func (n *P2PNode) GetListenAddr() string {
    return n.listenAddr
}

// HandleConnection handles incoming connections
func (n *P2PNode) handleConnection(conn net.Conn) {
	n.connMgr.AddConnection(conn)
	defer func() {
		n.connMgr.RemoveConnection(conn)
		conn.Close()
	}()

	for {
		var msg Message
		decoder := json.NewDecoder(conn)
		if err := decoder.Decode(&msg); err != nil {
			fmt.Printf("Failed to decode message: %v\n", err)
			return
		}

		switch MessageType(msg.Type) {
		case FileRequest:
			n.handleFileRequest(conn, msg.Data)
		case "ChunkRequest":
			n.handleChunkRequest(conn, msg.Data)
		}
	}
}

// HandleFileRequest handles file requests
func (n *P2PNode) handleFileRequest(conn net.Conn, data interface{}) {
	fileName, ok := data.(string)
	if !ok {
		fmt.Printf("Invalid file request data\n")
		return
	}

	// Read metadata
	metadata, err := n.storage.readMetadata(fileName)
	if err != nil {
		fmt.Printf("Failed to read metadata: %v\n", err)
		return
	}

	// Send metadata response
	response := NewMessage(FileResponse, metadata)
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(response); err != nil {
		fmt.Printf("Failed to send metadata response: %v\n", err)
		return
	}
}

// HandleChunkRequest handles chunk requests
func (n *P2PNode) handleChunkRequest(conn net.Conn, data interface{}) {
	request, ok := data.(ChunkRequest)
	if !ok {
		fmt.Printf("Invalid chunk request data\n")
		return
	}

	// Read chunk data
	chunkData, err := n.storage.readChunk(request.Hash)
	if err != nil {
		fmt.Printf("Failed to read chunk: %v\n", err)
		return
	}

	// Send chunk response
	chunkDataBytes, ok := chunkData.([]byte)
	if !ok {
		fmt.Printf("Invalid chunk data type\n")
		return
	}
	response := NewMessage(FileResponse, ChunkResponse{
		Hash: request.Hash,
		Data: chunkDataBytes,
	})
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(response); err != nil {
		fmt.Printf("Failed to send chunk response: %v\n", err)
		return
	}
}

// RequestFile requests a file from a peer
func (n *P2PNode) RequestFile(peerAddr string, fileName string) error {
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}
	defer conn.Close()

	// Send file request
	request := NewMessage(FileRequest, fileName)
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(request); err != nil {
		return fmt.Errorf("failed to send file request: %v", err)
	}

	// Receive metadata response
	var response Message
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&response); err != nil {
		return fmt.Errorf("failed to receive metadata response: %v", err)
	}

	// Parse metadata
	var metadata FileMetadata
	metadataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %v", err)
	}

	// Request each chunk
	for _, hash := range metadata.ChunkHashes {
		if err := n.requestChunk(peerAddr, hash); err != nil {
			return fmt.Errorf("failed to request chunk %s: %v", hash, err)
		}
	}

	return nil
}

// RequestChunk requests a chunk from a peer
func (n *P2PNode) requestChunk(peerAddr string, hash string) error {
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}
	defer conn.Close()
	// Send chunk request
	request := NewMessage("ChunkRequest", ChunkRequest{Hash: hash})
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(request); err != nil {
		fmt.Printf("Failed to send chunk request: %v\n", err)
		
	}

	// Receive chunk response
	var response Message
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&response); err != nil {
		return fmt.Errorf("failed to receive chunk response: %v", err)
	}

	// Parse chunk response
	var chunkResponse ChunkResponse
	chunkBytes, err := json.Marshal(response.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal chunk response: %v", err)
	}
	if err := json.Unmarshal(chunkBytes, &chunkResponse); err != nil {
		return fmt.Errorf("failed to unmarshal chunk response: %v", err)
	}

	// Store chunk
	if err := n.storage.storeChunk(hash, chunkResponse.Data); err != nil {
		return fmt.Errorf("failed to store chunk: %v", err)
	}

	return nil
}
