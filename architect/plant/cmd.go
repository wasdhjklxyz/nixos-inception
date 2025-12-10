// Package plant manages ISO generation and... (TODO)
package plant

import (
	"fmt"

	"github.com/wasdhjklxyz/nixos-inception/architect/crypto"
)

func ExecuteCmd(args []string) error {
	flags := parseArgs(args)

	certs, err := crypto.GenerateCertificates(flags.certDuration, flags.certSkew)
	if err != nil {
		return err
	}

	dir, err := WriteClientCredentials(certs)
	if err != nil {
		return err
	}
	fmt.Println(dir) // For consumption

	return nil
}
