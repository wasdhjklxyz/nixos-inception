package main

import (
	"fmt"
	"os"
)

func main() {
	_, err := newConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config: %v", err)
	}
}
