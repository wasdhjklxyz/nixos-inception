// Package crypto
package crypto

import (
	"fmt"
	"os"

	"filippo.io/age"
)

type AgeKeyPair struct {
	identity  *age.X25519Identity
	recipient *age.X25519Recipient
}

func ParseAgeIdentityFile(path string) ([]*AgeKeyPair, error) {
	in, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	ids, err := age.ParseIdentities(in)
	if err != nil {
		return nil, err
	}

	var keys []*AgeKeyPair
	for _, id := range ids {
		id, ok := id.(*age.X25519Identity)
		if !ok {
			continue
		}
		keys = append(keys, &AgeKeyPair{identity: id, recipient: id.Recipient()})
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no X25519 keys found in '%s'", path)
	}

	return keys, nil
}
