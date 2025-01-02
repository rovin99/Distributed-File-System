package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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



// Add a helper function to send messages reliably
func sendMessage(conn net.Conn, msg *Message) error {
    // Set write deadline
    if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
        return fmt.Errorf("failed to set write deadline: %v", err)
    }
    
    // Create an encoder for the connection
    encoder := json.NewEncoder(conn)
    if err := encoder.Encode(msg); err != nil {
        return fmt.Errorf("failed to encode message: %v", err)
    }
    
    return nil
}

// Add a helper function to receive messages reliably
func receiveMessage(conn net.Conn) (*Message, error) {
    // Set read deadline
    if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
        return nil, fmt.Errorf("failed to set read deadline: %v", err)
    }
    
    var msg Message
    decoder := json.NewDecoder(conn)
    if err := decoder.Decode(&msg); err != nil {
        if err == io.EOF {
            return nil, fmt.Errorf("connection closed by peer")
        }
        return nil, fmt.Errorf("failed to decode message: %v", err)
    }
    
    return &msg, nil
}
// HandleConnection handles incoming connections
func (n *P2PNode) handleConnection(conn net.Conn) {
    n.connMgr.AddConnection(conn)
    defer func() {
        n.connMgr.RemoveConnection(conn)
        conn.Close()
    }()

    for {
        msg, err := receiveMessage(conn)
        if err != nil {
            if !strings.Contains(err.Error(), "timeout") {
                fmt.Printf("Connection handler error: %v\n", err)
            }
            return
        }

        switch MessageType(msg.Type) {
        case FileRequest:
            var fileName string
            if err := json.Unmarshal(msg.Data, &fileName); err != nil {
                fmt.Printf("Failed to unmarshal file request: %v\n", err)
                continue
            }
            n.handleFileRequest(conn, fileName)
            
        case "ChunkRequest":
            var request ChunkRequest
            if err := json.Unmarshal(msg.Data, &request); err != nil {
                fmt.Printf("Failed to unmarshal chunk request: %v\n", err)
                continue
            }
            n.handleChunkRequest(conn, request)
        }
    }
}

// HandleFileRequest handles file requests
func (n *P2PNode) handleFileRequest(conn net.Conn, fileName string) {
    // Read metadata
    metadata, err := n.storage.readMetadata(fileName)
    if err != nil {
        fmt.Printf("Failed to read metadata: %v\n", err)
        return
    }

    // Send metadata response
    response := NewMessage(FileResponse, metadata)
    if err := sendMessage(conn, response); err != nil {
        fmt.Printf("Failed to send metadata response: %v\n", err)
        return
    }
}

// HandleChunkRequest handles chunk requests
func (n *P2PNode) handleChunkRequest(conn net.Conn, request ChunkRequest) {
    // Validate hash
    if request.Hash == "" {
        fmt.Printf("Empty hash in chunk request\n")
        return
    }

    // Read chunk data
    chunkPath := filepath.Join(n.storage.basePath, request.Hash)
    chunkData, err := os.ReadFile(chunkPath)
    if err != nil {
        fmt.Printf("Failed to read chunk %s: %v\n", request.Hash, err)
        return
    }

    // Send chunk response
    response := NewMessage(FileResponse, ChunkResponse{
        Hash: request.Hash,
        Data: chunkData,
    })

    if err := sendMessage(conn, response); err != nil {
        fmt.Printf("Failed to send chunk response: %v\n", err)
        return
    }
}

// Update the requestChunk method to handle the response properly
func (n *P2PNode) requestChunk(peerAddr string, hash string) error {
    conn, err := net.Dial("tcp", peerAddr)
    if err != nil {
        return fmt.Errorf("failed to connect to peer: %v", err)
    }
    defer conn.Close()

    // Create and send chunk request
    request := NewMessage("ChunkRequest", ChunkRequest{Hash: hash})
    if err := sendMessage(conn, request); err != nil {
        return fmt.Errorf("failed to send chunk request: %v", err)
    }

    // Receive chunk response
    response, err := receiveMessage(conn)
    if err != nil {
        return fmt.Errorf("failed to receive chunk response: %v", err)
    }

    // Parse chunk response
    var chunkResponse ChunkResponse
    if err := json.Unmarshal(response.Data, &chunkResponse); err != nil {
        return fmt.Errorf("failed to unmarshal chunk response: %v", err)
    }

    // Verify hash matches
    if chunkResponse.Hash != hash {
        return fmt.Errorf("received chunk hash %s doesn't match requested hash %s", 
            chunkResponse.Hash, hash)
    }

    // Store chunk
    if err := n.storage.storeChunk(hash, chunkResponse.Data); err != nil {
        return fmt.Errorf("failed to store chunk: %v", err)
    }

    return nil
}

func (n *P2PNode) verifyChunk(hash string) error {
    chunkPath := filepath.Join(n.storage.basePath, hash)
    _, err := os.Stat(chunkPath)
    if err != nil {
        return fmt.Errorf("chunk %s not found: %v", hash, err)
    }
    return nil
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
    if err := sendMessage(conn, request); err != nil {
        return fmt.Errorf("failed to send file request: %v", err)
    }

    // Receive metadata response
    response, err := receiveMessage(conn)
    if err != nil {
        return fmt.Errorf("failed to receive metadata response: %v", err)
    }

    // Parse metadata
    var metadata FileMetadata
    if err := json.Unmarshal(response.Data, &metadata); err != nil {
        return fmt.Errorf("failed to unmarshal metadata: %v", err)
    }

    // Request each chunk using a new connection
    for _, hash := range metadata.ChunkHashes {
        if err := n.requestChunk(peerAddr, hash); err != nil {
            return fmt.Errorf("failed to request chunk %s: %v", hash, err)
        }
    }

    return nil
}

