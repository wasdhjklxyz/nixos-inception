// Package crypto
package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
)

type RSAKeyPair struct {
	pub  *rsa.PublicKey
	priv *rsa.PrivateKey
}

func GenerateRSAKeyPair() (*RSAKeyPair, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %v", err)
	}

	return &RSAKeyPair{
		pub:  &priv.PublicKey,
		priv: priv,
	}, nil
}
