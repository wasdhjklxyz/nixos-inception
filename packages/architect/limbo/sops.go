// Package limbo...(TODO)
package limbo

/* FIXME: This entire project needs refactor (this shit should be in crypto) */

import (
	"fmt"

	"github.com/Mic92/ssh-to-age"
	"gopkg.in/yaml.v3"
)

func addSopsKey(key string) error {
	_, err := agessh.SSHPublicKeyToAge([]byte(key))
	if err != nil {
		return fmt.Errorf("failed to convert SSH public key to age: %v", err)
	}
	//yaml.NewDecoder()
	return nil
}
