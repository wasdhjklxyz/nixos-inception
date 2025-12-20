package main

import (
	"os"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/crypto"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/dream"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/limbo"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
)

func main() {
	flags := parseArgs(os.Args[1:])

	pipe, err := NewPipe(flags.ctlPipe)
	if err != nil {
		log.Error("failed to open control pipe: %v", err)
		os.Exit(1)
	}
	defer pipe.Close()

	log.Info("generating certificates...")
	certs, err := crypto.GenerateCertificates(flags.certDuration, flags.certSkew)
	if err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}

	log.Info("writing dreamer credentials...")
	dir, err := dream.WriteDreamerCredentials(certs)
	if err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}

	if err := pipe.Send(dir); err != nil {
		log.Error("pipe send failed: %v", err)
		os.Exit(1)
	}

	log.Info("loading closure...")
	c, err := limbo.NewClosure(flags.topLevel, flags.closure)
	if err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}

	if _, err := pipe.Recv(); err != nil {
		log.Error("pipe recv failed: %v", err)
		os.Exit(1)
	}

	log.Info("starting server...")
	if err := limbo.Descend(certs, flags.lport, c); err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}
}
