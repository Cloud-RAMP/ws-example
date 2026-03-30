package benchmark_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/Cloud-RAMP/ws-example.git/internal/server"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// BenchmarkServer-8           4011            290693 ns/op           20504 B/op     127 allocs/op
func BenchmarkServerNewClients(b *testing.B) {
	server.LogLevel = server.None

	// Make sure the server terminates when the context of the test expires
	ctx := b.Context()
	go server.Start(ctx)

	time.Sleep(500 * time.Millisecond)

	// Create client, send message, receive message
	for b.Loop() {
		func() {
			conn, _, _, err := ws.Dialer{}.Dial(ctx, "ws://localhost:8080")
			if err != nil {
				b.Errorf("Failed to connect to WebSocket server: %v\n", err)
				return
			}
			defer conn.Close()

			message := []byte("Hello, WebSocket server!")
			err = wsutil.WriteClientMessage(conn, ws.OpText, message)
			if err != nil {
				b.Errorf("Failed to send message: %v\n", err)
				return
			}

			messages, err := wsutil.ReadServerMessage(conn, nil)
			if err != nil {
				b.Errorf("Failed to recieve message: %v\n", err)
				return
			}

			for _, msg := range messages {
				if !bytes.Equal(msg.Payload, message) {
					b.Errorf("Message and response not equal")
					return
				}
			}
		}()
	}
}

// BenchmarkServerClientOne-8         39988             28555 ns/op            1168 B/op         16 allocs/op
func BenchmarkServerOneClient(b *testing.B) {
	server.LogLevel = server.None

	// Make sure the server terminates when the context of the test expires
	ctx := b.Context()
	go server.Start(ctx)

	time.Sleep(500 * time.Millisecond)

	// Create client, send message, receive message
	conn, _, _, err := ws.Dialer{}.Dial(ctx, "ws://localhost:8080")
	if err != nil {
		b.Errorf("Failed to connect to WebSocket server: %v\n", err)
		return
	}
	defer conn.Close()

	for b.Loop() {
		message := []byte("Hello, WebSocket server!")
		err = wsutil.WriteClientMessage(conn, ws.OpText, message)
		if err != nil {
			b.Errorf("Failed to send message: %v\n", err)
			continue
		}

		messages, err := wsutil.ReadServerMessage(conn, nil)
		if err != nil {
			b.Errorf("Failed to recieve message: %v\n", err)
			continue
		}

		for _, msg := range messages {
			if !bytes.Equal(msg.Payload, message) {
				b.Errorf("Message and response not equal\n")
				continue
			}
		}
	}
}

var NUM_CONNECTIONS = 5000

func BenchmarkServerNConnections(b *testing.B) {
	server.LogLevel = server.None

	ctx := b.Context()
	go server.Start(ctx)

	time.Sleep(500 * time.Millisecond)

	if NUM_CONNECTIONS <= 0 {
		b.Fatalf("NUM_CONNECTIONS must be > 0, got %d", NUM_CONNECTIONS)
	}

	clients := make([]net.Conn, 0, NUM_CONNECTIONS)
	for i := 0; i < NUM_CONNECTIONS; i++ {
		conn, _, _, err := ws.Dialer{}.Dial(ctx, "ws://localhost:8080")
		if err != nil {
			b.Fatalf("Failed to connect client %d: %v", i, err)
		}
		clients = append(clients, conn)
	}
	defer func() {
		for _, c := range clients {
			_ = c.Close()
		}
	}()

	message := []byte("Hello, WebSocket server!")
	b.ResetTimer()

	ops := 0
	for b.Loop() {
		for i, conn := range clients {
			err := wsutil.WriteClientMessage(conn, ws.OpText, message)
			if err != nil {
				b.Fatalf("Failed to send message from client %d: %v", i, err)
			}

			messages, err := wsutil.ReadServerMessage(conn, nil)
			if err != nil {
				b.Fatalf("Failed to recieve message for client %d: %v", i, err)
			}

			for _, msg := range messages {
				if !bytes.Equal(msg.Payload, message) {
					b.Fatalf("Message and response not equal for client %d", i)
				}
			}
			ops += 1
		}
	}

	b.ReportMetric(float64(ops)/float64(b.Elapsed().Seconds()), "ops/s")
}
