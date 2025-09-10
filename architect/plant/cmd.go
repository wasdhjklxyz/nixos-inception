// Package plant manages ISO generation and... (TODO)
package plant

import (
	"fmt"

	"github.com/wasdhjklxyz/nixos-inception/architect/crypto"
)

func ExecuteCmd(args []string) error {
	flags := parseArgs(args)

	caCertConfig := crypto.CertificateConfig{
		Name:     "nixos-inception",
		Duration: flags.certDuration,
		Skew:     flags.certSkew,
	}

	caCert, err := crypto.CreateCACertificate(caCertConfig)
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
