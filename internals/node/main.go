package main

import (
	"fmt"
	"log"
	
)

func main() {
	// Initialize storage and routing table
	InitStorage()
	InitRoutingTable()

	// Start server for peer communication
	go func() {
		err := StartServer(":8080") // Listening on port 8080
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Simulate joining the network
	bootstrapPeer := "http://localhost:8081" // Change as needed
	err := JoinNetwork(bootstrapPeer)
	if err != nil {
		log.Printf("Failed to join network: %v", err)
	} else {
		fmt.Println("Successfully joined the network!")
	}

	select {} // Keep the main function running
}
