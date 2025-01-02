package main

import (
	"bytes"
	"crypto/rand"
	"net"
	"os"
	"path/filepath"
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
	// Create test data
	testData := make([]byte, 2*1024*1024) // 2MB test file
	if _, err := rand.Read(testData); err != nil {
		t.Fatalf("Failed to generate test data: %v", err)
	}

	// Create test file
	testFile := filepath.Join(os.TempDir(), "test_transfer.dat")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Create two nodes
	node1, err := NewP2PNode("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create first node: %v", err)
	}
	if err := node1.Start(); err != nil {
		t.Fatalf("Failed to start first node: %v", err)
	}
	defer node1.Stop()

	node2, err := NewP2PNode("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create second node: %v", err)
	}
	if err := node2.Start(); err != nil {
		t.Fatalf("Failed to start second node: %v", err)
	}
	defer node2.Stop()

	t.Logf("Node1 listening on: %s", node1.listenAddr)
	t.Logf("Node2 listening on: %s", node2.listenAddr)

	// Split and store file on first node
	metadata, err := node1.storage.SplitFile(testFile)
	if err != nil {
		t.Fatalf("Failed to split file: %v", err)
	}

	// Request file from second node
	err = node2.RequestFile(node1.listenAddr, metadata.FileName)
	if err != nil {
		t.Fatalf("Failed to request file: %v", err)
	}

	// Give some time for transfer to complete
	time.Sleep(time.Second)

	// Reassemble file
	outputPath := filepath.Join(os.TempDir(), "received_"+metadata.FileName)
	if err := node2.storage.ReassembleFile(metadata, outputPath); err != nil {
		t.Fatalf("Failed to reassemble file: %v", err)
	}
	defer os.Remove(outputPath)

	// Verify content
	receivedData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read received file: %v", err)
	}

	if !bytes.Equal(testData, receivedData) {
		t.Error("Received file content doesn't match original")
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