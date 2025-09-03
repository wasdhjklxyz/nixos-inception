package main

import (
	"fmt"
	"os"
)

func main() {
	cfg, err := newConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config: %v", err)
	}

	srv, err := newServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v", err)
	}

	srv.run()
}
