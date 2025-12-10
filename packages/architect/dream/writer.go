// Package dream manages ISO generation and... (TODO)
package dream

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/crypto"
)

func writePEM(path, pemType string, der []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: pemType, Bytes: der})
}

func WriteClientCredentials(certs *crypto.Certificates) (string, error) {
	dir, err := os.MkdirTemp("", "nixos-inception-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	certPath := filepath.Join(dir, "client.crt")
	if err := writePEM(certPath, "CERTIFICATE", certs.ClientCertDER); err != nil {
		return "", err
	}

	key, err := x509.MarshalPKCS8PrivateKey(certs.ClientKeyPair.Priv)
	if err != nil {
		return "", err
	}

	keyPath := filepath.Join(dir, "client.key")

	if err := writePEM(keyPath, "PRIVATE KEY", key); err != nil {
		return "", err
	}

	return dir, nil
}
