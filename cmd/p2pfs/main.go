package main

import (
    "log"
    "github.com/rovin99/Distributed-File-System/internals/node"
	"github.com/rovin99/Distributed-File-System/internals/network"
)

func main() {
    communicationService := network.NewCommunicationService()
    
    nodeInstance := node.NewNode(nil)
    
    if err := nodeInstance.Start(); err != nil {
        log.Fatalf("Failed to start node: %v", err)
    }

    if err := communicationService.Start(":8080"); err != nil {
        log.Fatalf("Failed to start communication service: %v", err)
    }
}