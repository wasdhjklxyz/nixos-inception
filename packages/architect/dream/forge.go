// Package dream manages ISO generation and... (TODO)
package dream

import (
	"fmt"
	"time"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/crypto"
)

func Forge(certDuration, certSkew time.Duration) error {
	certs, err := crypto.GenerateCertificates(certDuration, certSkew)
	if err != nil {
		return err
	}

	dir, err := WriteDreamerCredentials(certs)
	if err != nil {
		return err
	}
	fmt.Println(dir) // For consumption

	return nil
}
