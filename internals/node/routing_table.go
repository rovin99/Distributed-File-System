package main

import (
	"fmt"
	"sync"
)

var (
	routingTable      = make(map[string]string) // Map of peerID to peer address
	routingTableMutex sync.RWMutex             // Mutex for thread-safe access
)

// InitRoutingTable initializes the routing table.
func InitRoutingTable() {
	fmt.Println("Routing table initialized.")
}

// AddPeer adds a peer to the routing table.
func AddPeer(peerID, address string) {
	routingTableMutex.Lock()
	defer routingTableMutex.Unlock()

	routingTable[peerID] = address
	fmt.Printf("Added peer: %s (%s)\n", peerID, address)
}

// GetPeer retrieves a peer address by its ID.
func GetPeer(peerID string) (string, error) {
	routingTableMutex.RLock()
	defer routingTableMutex.RUnlock()

	address, exists := routingTable[peerID]
	if !exists {
		return "", fmt.Errorf("peer not found: %s", peerID)
	}
	return address, nil
}
