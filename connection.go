
package main

import (
    "net"
    "sync"
)

// Connection represents a peer connection
type Connection struct {
    Conn net.Conn
	
   
}

// ConnectionManager manages peer connections
type ConnectionManager struct {
    connections map[string]*Connection
    mu          sync.RWMutex
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager() *ConnectionManager {
    return &ConnectionManager{
        connections: make(map[string]*Connection),
    }
}

// AddConnection adds a new connection
func (cm *ConnectionManager) AddConnection(conn net.Conn) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    cm.connections[conn.RemoteAddr().String()] = &Connection{
        Conn: conn,
    }
}

// RemoveConnection removes a connection
func (cm *ConnectionManager) RemoveConnection(conn net.Conn) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    delete(cm.connections, conn.RemoteAddr().String())
}

// GetConnection returns a connection by address
func (cm *ConnectionManager) GetConnection(addr string) *Connection {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    return cm.connections[addr]
}