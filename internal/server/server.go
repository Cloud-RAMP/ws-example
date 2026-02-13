package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// Define start method here so that we can use it in testing
func Start() {
	http.HandleFunc("/", handleConnection)
	http.ListenAndServe(":8080", nil)
}

// Connection handler function
func handleConnection(w http.ResponseWriter, r *http.Request) {
	// Updagrade the HTTP connection to WS
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		// handle error
		fmt.Println("Error upgrading protocols")
		return
	}

	// Here we can access info about the client, good for logging + billing purposes
	fmt.Println("New client connected:")
	fmt.Println("IP:", r.RemoteAddr)
	fmt.Println("Domain:", r.Host)

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
				fmt.Println("Client disconnected")
				return
			} else if err != nil {
				// handle error
				fmt.Println("error reading:", err)
				return
			}

			// Probably not super necessary, good safeguard though
			// Other operation types CAN be sent
			if op.IsData() {
				fmt.Println("Client message:", string(msg))
			}

			// Write a message to the client as the server (in this case, echo it)
			err = wsutil.WriteServerMessage(conn, op, msg)
			if err == io.EOF {
				fmt.Println("Client disconnected")
				return
			} else if err != nil {
				fmt.Println("error reading:", err)
				return
			}
		}
	}()
}
