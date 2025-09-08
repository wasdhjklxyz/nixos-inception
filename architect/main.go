package main

import (
	"fmt"
	"os"

	"github.com/wasdhjklxyz/nixos-inception/architect/limbo"
	"github.com/wasdhjklxyz/nixos-inception/architect/plant"
)

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	commands := map[string]func([]string) error{
		"plant": plant.ExecuteCmd,
		"limbo": limbo.ExecuteCmd,
	}

	cmd, exists := commands[os.Args[1]]
	if !exists {
		os.Exit(1)
	}

	if err := cmd(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
