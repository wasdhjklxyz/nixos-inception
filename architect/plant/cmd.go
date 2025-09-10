// Package plant manages ISO generation and... (TODO)
package plant

import (
	"fmt"

	"github.com/wasdhjklxyz/nixos-inception/architect/crypto"
)

func ExecuteCmd(args []string) error {
	flags := parseArgs(args)

	caCert, err := crypto.NewCACertificate(
		flags.certDuration,
		flags.certSkew,
	)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %v", err)
	}
	_ = caCert

	clientCert, err := crypto.NewClientCertificate(
		flags.certDuration,
		flags.certSkew,
	)
	if err != nil {
		return fmt.Errorf("failed to create client certificate: %v", err)
	}
	_ = clientCert

	keys, err := crypto.ParseAgeIdentityFile(flags.ageIdentityFile)
	if err != nil {
		return fmt.Errorf("failed to parse age keys: %v", err)
	}
	_ = keys

	return nil
}
