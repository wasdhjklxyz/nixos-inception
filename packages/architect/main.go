package main

import (
	"fmt"
	"os"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/dream"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/limbo"
)

func main() {
	flags := parseArgs(os.Args[1:])

	if err := dream.Forge(flags.certDuration, flags.certSkew); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	pipe, _ := os.Open(flags.ctlPipe)
	buf := make([]byte, 8)
	pipe.Read(buf)

	/* FIXME: Use port from nix */
	if err := limbo.StartHTTPListener(12345); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
