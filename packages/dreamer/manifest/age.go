// Package manifest...(TODO)
package manifest

import (
	"fmt"

	"filippo.io/age"
)

type AgeKeyPair struct {
	identity  *age.X25519Identity
	recipient *age.X25519Recipient
}

func generateAgeKeyPair() (*AgeKeyPair, error) {
	id, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("failed to generated X25519 identity: %v", err)
	}
	return &AgeKeyPair{
		identity:  id,
		recipient: id.Recipient(),
	}, nil
}
