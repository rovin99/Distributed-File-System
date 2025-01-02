package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// JoinNetwork connects to a bootstrap peer and adds it to the routing table.
func JoinNetwork(bootstrapPeer string) error {
	// Simulate adding the bootstrap peer
	peerID := "bootstrap"
	AddPeer(peerID, bootstrapPeer)
	fmt.Printf("Joined network via: %s\n", bootstrapPeer)
	return nil
}

// StoreChunkOnPeer sends a chunk to another peer for storage.
func StoreChunkOnPeer(peerAddress, chunkID string, data []byte) error {
	url := fmt.Sprintf("%s/store", peerAddress)
	reqBody, _ := json.Marshal(map[string]interface{}{
		"chunk_id": chunkID,
		"data":     data,
	})

	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to store chunk on peer: %v", err)
	}
	return nil
}
