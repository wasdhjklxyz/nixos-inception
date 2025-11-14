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
	_ = certs

	keys, err := crypto.ParseAgeIdentityFile(flags.ageIdentityFile)
	if err != nil {
		return fmt.Errorf("failed to parse age keys: %v", err)
	}
	_ = keys

	return nil
}
