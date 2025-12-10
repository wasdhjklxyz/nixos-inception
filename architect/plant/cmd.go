// Package plant manages ISO generation and... (TODO)
package plant

import (
	"fmt"
	"os"

	"github.com/wasdhjklxyz/nixos-inception/architect/crypto"
)

func ExecuteCmd(args []string) error {
	flags := parseArgs(args)

	certs, err := crypto.GenerateCertificates(flags.certDuration, flags.certSkew)
	if err != nil {
		return err
	}

	dir, err := WriteClientCredentials(certs)
	if err != nil {
		return err
	}
	fmt.Println(dir) // For consumption

	pipe, _ := os.Open(flags.ctlPipe)
	buf := make([]byte, 8)
	pipe.Read(buf)

	err = StartHTTPListener(12345) // FIXME: Use port from nix
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err
}
