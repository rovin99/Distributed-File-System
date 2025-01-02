package main

import (
	"encoding/json"
	"fmt"
	"net"
)

// MessageType represents the type of a message
type MessageType string

const (
    // Ping represents a ping message
    Ping MessageType = "ping"
    // Pong represents a pong message
    Pong MessageType = "pong"
    // FileRequest represents a file request message
    FileRequest MessageType = "file_request"
    // FileResponse represents a file response message
    FileResponse MessageType = "file_response"
)

// Message represents a basic message
type Message struct {
    Type string      `json:"type"`
    Data interface{} `json:"data"`
}

// NewMessage creates a new message
func NewMessage(t MessageType, data interface{}) *Message {
    return &Message{
        Type: string(t),
        Data: data,
    }
}

// Serialize serializes a message to a JSON string
func (m *Message) Serialize() ([]byte, error) {
    return json.Marshal(m)
}

// Deserialize deserializes a JSON string to a message
func Deserialize(data []byte) (*Message, error) {
    var msg Message
    err := json.Unmarshal(data, &msg)
    return &msg, err
}



// TCP Server
func startServer() {
    ln, err := net.Listen("tcp", ":8080")
    if err!= nil {
        fmt.Println(err)
        return
    }
    defer ln.Close()

    fmt.Println("Server listening on port 8080")

    for {
        conn, err := ln.Accept()
        if err!= nil {
            fmt.Println(err)
            continue
        }
        go handleConnection(conn)
    }
}

// TCP Client
func startClient() {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err!= nil {
        fmt.Println(err)
        return
    }
    defer conn.Close()

    // Send a ping message
    pingMsg := NewMessage(Ping, nil)
    pingData, err := pingMsg.Serialize()
    if err!= nil {
        fmt.Println(err)
        return
    }
    _, err = conn.Write(pingData)
    if err!= nil {
        fmt.Println(err)
        return
    }

    // Receive a response
    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err!= nil {
        fmt.Println(err)
        return
    }
    responseMsg, err := Deserialize(buf[:n])
    if err!= nil {
        fmt.Println(err)
        return
    }
    fmt.Println(responseMsg)
}

func handleConnection(conn net.Conn) {
    // Receive a message
    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err!= nil {
        fmt.Println(err)
        return
    }
    msg, err := Deserialize(buf[:n])
    if err!= nil {
        fmt.Println(err)
        return
    }

    // Handle the message
    switch msg.Type {
    case string(Ping):
        // Send a pong response
        pongMsg := NewMessage(Pong, nil)
        pongData, err := pongMsg.Serialize()
        if err!= nil {
            fmt.Println(err)
            return
        }
        _, err = conn.Write(pongData)
        if err!= nil {
            fmt.Println(err)
            return
        }
    default:
        fmt.Println("Unknown message type")
    }
}

func StartP2P() {
    go startServer()
    startClient()
}