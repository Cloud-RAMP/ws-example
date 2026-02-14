package main

import (
	"context"

	"github.com/Cloud-RAMP/ws-example.git/internal/server"
)

func main() {
	server.Start(context.Background())
}
