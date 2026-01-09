// Package limbo...(TODO)
package limbo

/* FIXME: This entire project needs refactor (this shit should be in crypto) */

import (
	"fmt"

	"github.com/Mic92/ssh-to-age"
)

func addSopsKey(key string) error {
	_, err := agessh.SSHPublicKeyToAge([]byte(key))
	if err != nil {
		return fmt.Errorf("failed to convert SSH key to age key: %v", err)
	}
	return nil
}
