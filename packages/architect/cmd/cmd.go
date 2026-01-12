// Package cmd...(TODO)
package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/crypto"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/dream"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/limbo"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
)

func Run(args []string) error {
	flags := parseArgs(args)

	log.Info("generating certificates...")
	certs, err := crypto.GenerateCertificates(flags.certDuration, flags.certSkew)
	if err != nil {
		return fmt.Errorf("failed to generate certificates: %v", err)
	}

	log.Info("writing dreamer credentials...")
	dir, err := dream.WriteDreamerCredentials(certs)
	if err != nil {
		return fmt.Errorf("failed to write dreamer credentials: %v", err)
	}

	fmt.Fprintln(os.Stdout, dir) /* NOTE: Specifying stdout since script requires */

	log.Info("loading closure...")
	c, err := limbo.NewClosure(flags.topLevel, flags.closure, flags.diskoScript, flags.diskoDevice, flags.diskSelection)
	if err != nil {
		return err
	}

	stdin := bufio.NewScanner(os.Stdin)
	if stdin.Scan() {
		log.Info("starting server...")
		if err := limbo.Descend(certs, flags.lport, c); err != nil {
			return err
		}
	}

	return nil
}
