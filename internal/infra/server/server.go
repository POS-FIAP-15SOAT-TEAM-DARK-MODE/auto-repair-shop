package server

import "context"

// Server defines the interface for all servers (HTTP, gRPC, etc.)
type Server interface {
	// Start begins the server's listener
	Start(ctx context.Context) error
	// Stop gracefully shuts down the server
	Stop(ctx context.Context) error
	// Name returns the name of the server for logging purposes
	Name() string
}
