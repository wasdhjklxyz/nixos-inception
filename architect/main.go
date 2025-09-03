package main

import (
	"fmt"
	"os"
)

func main() {
	cfg, err := newConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config: %v", err)
		os.Exit(1)
	}

	srv, err := newServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v", err)
		os.Exit(1)
	}

	srv.run()
}
