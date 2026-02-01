package main

import (
	"fmt"
	"os"

	"github.com/wasdhjklxyz/nixos-inception/packages/dreamer/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
