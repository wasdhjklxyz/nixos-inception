package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/crypto"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/dream"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/limbo"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
)

func main() {
	flags := parseArgs(os.Args[1:])

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

	fmt.Fprintln(os.Stdout, dir) /* NOTE: Specifying stdout since script requires */

	log.Info("loading closure...")
	c, err := limbo.NewClosure(flags.topLevel, flags.closure, flags.diskoScript, flags.diskoDevice, flags.diskSelection)
	if err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}

	stdin := bufio.NewScanner(os.Stdin)
	if stdin.Scan() {
		log.Info("starting server...")
		if err := limbo.Descend(certs, flags.lport, c); err != nil {
			log.Error("%v", err)
			os.Exit(1)
		}
	}
}
