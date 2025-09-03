package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"filippo.io/age"
)

type keyPair struct {
	identity  *age.X25519Identity
	recipient *age.X25519Recipient
}

type config struct {
	port int
	keys []keyPair
}

func newConfig() (*config, error) {
	defaultPort, err := net.LookupPort("tcp", "http")
	if err != nil {
		return nil, fmt.Errorf("failed to get port for HTTP service: %v", err)
	}

	var (
		port    = flag.Int("port", defaultPort, "Listen port")
		keyFile = flag.String("age-key", "", "Path to age identity file (required)")
		help    = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		flag.Usage()
	}

	if *keyFile == "" {
		flag.Usage()
		return nil, fmt.Errorf("missing required argument: age-key")
	}

	in, err := os.Open(*keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open key file: %v", err)
	}
	defer in.Close()

	ids, err := age.ParseIdentities(in)
	if err != nil {
		return nil, fmt.Errorf("failed to parse identities: %v", err)
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
		return nil, fmt.Errorf("no X25519 keys found")
	}

	return &config{port: *port, keys: keys}, nil
}
