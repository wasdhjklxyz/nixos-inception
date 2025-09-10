// Package plant manages ISO generation and... (TODO)
package plant

import (
	"fmt"

	"github.com/wasdhjklxyz/nixos-inception/architect/crypto"
)

func ExecuteCmd(args []string) error {
	flags := parseArgs(args)

	caCert, err := crypto.CreateCACertificate(flags.certDuration, flags.certSkew)
	if err != nil {
		return fmt.Errorf("failed to create certificate authority: %v", err)
	}
	_ = caCert

	keys, err := crypto.ParseAgeIdentityFile(flags.ageIdentityFile)
	if err != nil {
		return fmt.Errorf("failed to parse age keys: %v", err)
	}
	_ = keys

	return nil
}
