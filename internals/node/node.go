package node

import (
    "net"
    "sync"
    
)

type Node struct {
    ID         string
    Address    net.Addr
    RoutingTable *RoutingTable
    Storage    *LocalStorage
    mu         sync.RWMutex
}

func NewNode(address net.Addr) *Node {
    return &Node{
        ID:           generateNodeID(),
        Address:      address,
        RoutingTable: NewRoutingTable(),
        Storage:      NewLocalStorage(),
    }
}

func (n *Node) Start() error {
    // Initialize node, connect to network

	
    return nil
}
