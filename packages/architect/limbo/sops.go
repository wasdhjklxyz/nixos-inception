// Package limbo...(TODO)
package limbo

/* FIXME: This entire project needs refactor (this shit should be in crypto) */

import (
	"fmt"

	"github.com/Mic92/ssh-to-age"
)

func publicSSHToAge(key string) (*string, error) {
	ageKey, err := agessh.SSHPublicKeyToAge([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("failed to convert SSH public key to age: %v")
	}
	return ageKey, nil
}
