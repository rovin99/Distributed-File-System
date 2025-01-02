package main

import (
	"bytes"
	"crypto/rand"

	"net"
	"os"

	"testing"
	"time"
)

func TestNetworkBasics(t *testing.T) {
    // Create and start first node
    node1, err := NewP2PNode("127.0.0.1:0")
    if err != nil {
        t.Fatalf("Failed to create first node: %v", err)
    }
    if err := node1.Start(); err != nil {
        t.Fatalf("Failed to start first node: %v", err)
    }
    defer node1.Stop()

    // Get the actual address the node is listening on
    actualAddr := node1.GetListenAddr()
    
    // Test node is listening
    conn, err := net.Dial("tcp", actualAddr)
    if err != nil {
        t.Fatalf("Failed to connect to node: %v", err)
    }
    defer conn.Close()

    t.Logf("Successfully connected to node at %s", actualAddr)
}

func TestFileTransfer(t *testing.T) {
    // Create test file with sufficient content to trigger chunking
    testContent := make([]byte, 2*1024*1024) // 2MB to ensure multiple chunks
    rand.Read(testContent)                    // Fill with random data
    testFile := "test.txt"
    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to create test file: %v", err)
    }
    defer os.Remove(testFile)

    // Create and start nodes
    node1, err := NewP2PNode("127.0.0.1:0")
    if err != nil {
        t.Fatalf("Failed to create node1: %v", err)
    }
    defer node1.Stop()

    node2, err := NewP2PNode("127.0.0.1:0")
    if err != nil {
        t.Fatalf("Failed to create node2: %v", err)
    }
    defer node2.Stop()

    if err := node1.Start(); err != nil {
        t.Fatalf("Failed to start node1: %v", err)
    }
    if err := node2.Start(); err != nil {
        t.Fatalf("Failed to start node2: %v", err)
    }

    t.Logf("Node1 listening on: %s", node1.GetListenAddr())
    t.Logf("Node2 listening on: %s", node2.GetListenAddr())

    // Split file on node1 and log metadata
    metadata, err := node1.storage.SplitFile(testFile)
    if err != nil {
        t.Fatalf("Failed to split file: %v", err)
    }

    t.Logf("File split into %d chunks", len(metadata.ChunkHashes))
    for i, hash := range metadata.ChunkHashes {
        t.Logf("Chunk %d: %s (size: %d)", i, hash, metadata.ChunkSizes[i])
    }

    // Add delay for node readiness
    time.Sleep(100 * time.Millisecond)

    // Request file transfer
    if err := node2.RequestFile(node1.GetListenAddr(), testFile); err != nil {
        t.Fatalf("Failed to request file: %v", err)
    }

    // Add delay for transfer completion
    time.Sleep(500 * time.Millisecond)

    // Verify chunks were received
    for i, hash := range metadata.ChunkHashes {
        if err := node2.verifyChunk(hash); err != nil {
            t.Errorf("Chunk %d verification failed: %v", i, err)
        }
    }

    // Reassemble and verify file
    outputPath := "reassembled_" + testFile
    if err := node2.storage.ReassembleFile(metadata, outputPath); err != nil {
        t.Fatalf("Failed to reassemble file: %v", err)
    }
    defer os.Remove(outputPath)

    // Verify content
    received, err := os.ReadFile(outputPath)
    if err != nil {
        t.Fatalf("Failed to read reassembled file: %v", err)
    }

    if !bytes.Equal(received, testContent) {
        t.Error("Reassembled content doesn't match original")
    } else {
        t.Log("File transfer and reassembly successful")
    }
}

func TestConnectionManager(t *testing.T) {
	cm := NewConnectionManager()

	// Create test listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Create test connection
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	// Test connection management
	cm.AddConnection(conn)
	if c := cm.GetConnection(conn.RemoteAddr().String()); c == nil {
		t.Error("Failed to retrieve added connection")
	}

	cm.RemoveConnection(conn)
	if c := cm.GetConnection(conn.RemoteAddr().String()); c != nil {
		t.Error("Connection still exists after removal")
	}
}