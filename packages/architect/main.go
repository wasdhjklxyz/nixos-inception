package main

import (
	"os"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/cmd"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
)

func main() {
	if err := cmd.Run(os.Args[1:]); err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}
}
