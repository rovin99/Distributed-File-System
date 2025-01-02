package main

import (
	"encoding/json"
	
	"net/http"
)

// StartServer starts the HTTP server for peer communication.
func StartServer(address string) error {
	http.HandleFunc("/store", handleStoreChunk)
	http.HandleFunc("/retrieve", handleRetrieveChunk)
	return http.ListenAndServe(address, nil)
}

// handleStoreChunk handles storing a chunk on this peer.
func handleStoreChunk(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChunkID string `json:"chunk_id"`
		Data    []byte `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := StoreChunk(req.ChunkID, req.Data)
	if err != nil {
		http.Error(w, "Failed to store chunk", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// handleRetrieveChunk handles retrieving a chunk from this peer.
func handleRetrieveChunk(w http.ResponseWriter, r *http.Request) {
	chunkID := r.URL.Query().Get("chunk_id")
	if chunkID == "" {
		http.Error(w, "Missing chunk_id", http.StatusBadRequest)
		return
	}

	data, err := GetChunk(chunkID)
	if err != nil {
		http.Error(w, "Chunk not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}
