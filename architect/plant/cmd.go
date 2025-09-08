// Package plant manages ISO generation and... (TODO)
package plant

import (
	"flag"
	"fmt"
	"os"
	"path"

	"filippo.io/age"
)

type keyPair struct {
	identity  *age.X25519Identity
	recipient *age.X25519Recipient
}

func getKeys(path string) ([]keyPair, error) {
	in, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	ids, err := age.ParseIdentities(in)
	if err != nil {
		return nil, err
	}

	var keys []keyPair
	for _, id := range ids {
		id, ok := id.(*age.X25519Identity)
		if !ok {
			continue
		}
		keys = append(keys, keyPair{identity: id, recipient: id.Recipient()})
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no X25519 keys found in '%s'", path)
	}

	return keys, nil
}

func ExecuteCmd(args []string) error {
	fs := flag.NewFlagSet("plant", flag.ExitOnError)
	keyFile := fs.String("age-key", "", "Path to age identity file (required)")
	fs.Parse(args)

	if *keyFile == "" {
		fs.Usage()
	}

	_, err := getKeys(path.Clean(*keyFile))
	if err != nil {
		return err
	}

	return nil
}
