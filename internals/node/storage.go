package main

import (
	"fmt"
	
	"sync"
)

var (
	storage      = make(map[string][]byte) // Map to store chunks (chunkID -> chunk data)
	storageMutex sync.RWMutex              // Mutex for thread-safe access
)

// InitStorage initializes the storage system.
func InitStorage() {
	fmt.Println("Storage initialized.")
}

// StoreChunk saves a chunk to local storage.
func StoreChunk(chunkID string, data []byte) error {
	storageMutex.Lock()
	defer storageMutex.Unlock()

	storage[chunkID] = data
	fmt.Printf("Stored chunk: %s\n", chunkID)
	return nil
}

// GetChunk retrieves a chunk from local storage.
func GetChunk(chunkID string) ([]byte, error) {
	storageMutex.RLock()
	defer storageMutex.RUnlock()

	data, exists := storage[chunkID]
	if !exists {
		return nil, fmt.Errorf("chunk not found: %s", chunkID)
	}
	return data, nil
}
