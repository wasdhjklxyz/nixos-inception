// Package crypto
package crypto

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

func generateSerialNumber() (*big.Int, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	b[0] &= 0x7F // Force positive (MSB = 0)
	return big.NewInt(0).SetBytes(b), nil
}

func CreateCACertificate(dur, skew time.Duration) (*x509.Certificate, error) {
	now := time.Now().UTC()

	sn, err := generateSerialNumber()
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber: sn,
		Subject:      pkix.Name{CommonName: "nixos-inception"},

		NotBefore: now.Add(-skew),
		NotAfter:  now.Add(dur),

		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,

		KeyUsage: x509.KeyUsageCertSign |
			x509.KeyUsageCRLSign |
			x509.KeyUsageDigitalSignature,

		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}, nil
}
