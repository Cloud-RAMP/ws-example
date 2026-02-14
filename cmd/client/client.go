package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// Runs a simple client that connects to the web socket server and sends / receives messages
func main() {
	// Create a context with a timeout. Cancel dial if not met within 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to ws server
	conn, _, _, err := ws.Dialer{}.Dial(ctx, "ws://localhost:8080")
	if err != nil {
		fmt.Printf("Failed to connect to WebSocket server: %v\n", err)
	}
	defer conn.Close()

	// Send a message to the server
	message := []byte("Hello, WebSocket server!")
	err = wsutil.WriteClientMessage(conn, ws.OpText, message)
	if err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
	}

	// Read some amount of messages back from the server
	// The nil parameter represents a pre-allocated buffer we can store the messages in.
	// We could use this for performance, but it restricts the total message length
	messages, err := wsutil.ReadServerMessage(conn, nil)
	if err != nil {
		fmt.Printf("Failed to recieve message: %v\n", err)
	}

	// Read messages that came back from the server
	for _, msg := range messages {
		msgString := string(msg.Payload)
		fmt.Println("Response from server:", msgString)
	}

	// Example delay
	time.Sleep(1 * time.Second)

	// Repeat process again
	message = []byte("Sending second message")
	err = wsutil.WriteClientMessage(conn, ws.OpText, message)
	if err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
	}

	messages, err = wsutil.ReadServerMessage(conn, nil)
	if err != nil {
		fmt.Printf("Failed to recieve message: %v\n", err)
	}

	for _, msg := range messages {
		msgString := string(msg.Payload)
		fmt.Println("Response from server:", msgString)
	}
}
