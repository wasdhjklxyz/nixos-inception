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

func WriteDreamerCredentials(certs *crypto.Certificates) (string, error) {
	dir, err := os.MkdirTemp("", "nixos-inception-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	caPath := filepath.Join(dir, "ca.crt")
	if err := writePEM(caPath, "CERTIFICATE", certs.CACertDER); err != nil {
		return "", err
	}

	certPath := filepath.Join(dir, "dreamer.crt")
	if err := writePEM(certPath, "CERTIFICATE", certs.DreamerCertDER); err != nil {
		return "", err
	}

	key, err := x509.MarshalPKCS8PrivateKey(certs.DreamerKeyPair.Priv)
	if err != nil {
		return "", err
	}

	keyPath := filepath.Join(dir, "dreamer.key")

	if err := writePEM(keyPath, "PRIVATE KEY", key); err != nil {
		return "", err
	}

	return dir, nil
}
