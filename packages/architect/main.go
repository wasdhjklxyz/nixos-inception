package main

import (
	"fmt"
	"os"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/crypto"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/dream"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/limbo"
)

func main() {
	flags := parseArgs(os.Args[1:])

	certs, err := crypto.GenerateCertificates(flags.certDuration, flags.certSkew)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dir, err := dream.WriteDreamerCredentials(certs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(dir) // For consumption

	pipe, _ := os.Open(flags.ctlPipe)
	buf := make([]byte, 8)
	pipe.Read(buf)

	if err := limbo.Descend(certs, flags.lport); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
