package server

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type logLevel string

const (
	Verbose logLevel = "verbose"
	None    logLevel = "none"
)

var LogLevel = Verbose

// Wrapper function to log only if logging is turned on.
// Useful for testing high performance
func ServerLog(args ...any) {
	if LogLevel == Verbose {
		fmt.Println(args...)
	}
}

// Define start method here so that we can use it in testing
//
// It's a good habit to use these "ctx" objects, since they give us fine grained
// control over server state
func Start(ctx context.Context) {
	server := &http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(handleConnection),
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			fmt.Println("Error starting server", err)
			return
		}
	}()
	ServerLog("Server started!")

	// This is "channel" syntax, basically it means that we wait until something is sent in the channel
	// Channels can be used to send data, but in this case it is used as a signal by sending an empty struct
	<-ctx.Done()
	ServerLog("Shutting down...")
	server.Shutdown(ctx)
}

// Connection handler function
func handleConnection(w http.ResponseWriter, r *http.Request) {
	// Updagrade the HTTP connection to WS
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		ServerLog("Error upgrading protocols")
		return
	}

	// Here we can access info about the client, good for logging + billing purposes
	ServerLog("New client connected:")
	ServerLog("IP:", r.RemoteAddr)
	ServerLog("Domain:", r.Host)

	// Start separate goroutine for each connection
	// This does not scale super well, think about maybe worker pools or something
	go func() {
		defer conn.Close() // when the function returns, close the connection

		// infinite loop
		for {
			// Read data from the client on the connection
			// see https://datatracker.ietf.org/doc/html/rfc6455#section-5.5 for info on "op"
			msg, op, err := wsutil.ReadClientData(conn)
			if err == io.EOF {
				ServerLog("Client disconnected")
				return
			} else if err != nil {
				// handle error
				ServerLog("error reading:", err)
				return
			}

			// Probably not super necessary, good safeguard though
			// Other operation types CAN be sent
			if op.IsData() {
				ServerLog("Client message:", string(msg))
			}

			// Write a message to the client as the server (in this case, echo it)
			err = wsutil.WriteServerMessage(conn, op, msg)
			if err == io.EOF {
				ServerLog("Client disconnected")
				return
			} else if err != nil {
				ServerLog("error reading:", err)
				return
			}
		}
	}()
}
