// Package crypto
package crypto

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"
)

type Certificates struct {
	CAKeyPair     *RSAKeyPair
	CACertDER     []byte
	ClientKeyPair *RSAKeyPair
	ClientCertDER []byte
}

const commonName = "nixos-inception"

func generateSerialNumber() (*big.Int, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	b[0] &= 0x7F // Force positive (MSB = 0)
	return big.NewInt(0).SetBytes(b), nil
}

func createCATemplate(dur, skew time.Duration) (*x509.Certificate, error) {
	now := time.Now().UTC()

	sn, err := generateSerialNumber()
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber: sn,
		Subject:      pkix.Name{CommonName: commonName + "-CA"},

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

func createClientTemplate(dur, skew time.Duration) (*x509.Certificate, error) {
	now := time.Now().UTC()

	sn, err := generateSerialNumber()
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber: sn,
		Subject:      pkix.Name{CommonName: commonName + "-client"},

		NotBefore: now.Add(-skew),
		NotAfter:  now.Add(dur),

		KeyUsage: x509.KeyUsageDigitalSignature,

		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}, nil
}

func GenerateCertificates(dur, skew time.Duration) (*Certificates, error) {
	caKeyPair, err := GenerateRSAKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate CA RSA key pair: %v", err)
	}

	caCertTemplate, err := createCATemplate(dur, skew)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate template: %v", err)
	}

	caCertDER, err := x509.CreateCertificate(
		rand.Reader,
		caCertTemplate,
		caCertTemplate,
		caKeyPair.pub,
		caKeyPair.Priv,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate: %v", err)
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate DER: %v", err)
	}

	clientKeyPair, err := GenerateRSAKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client RSA key pair: %v", err)
	}

	clientCertTemplate, err := createClientTemplate(dur, skew)
	if err != nil {
		return nil, fmt.Errorf("failed to create client certificate template: %v", err)
	}

	clientCertDER, err := x509.CreateCertificate(
		rand.Reader,
		clientCertTemplate,
		caCert,
		clientKeyPair.pub,
		caKeyPair.Priv,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client certificate: %v", err)
	}

	return &Certificates{
		CAKeyPair:     caKeyPair,
		CACertDER:     caCertDER,
		ClientKeyPair: clientKeyPair,
		ClientCertDER: clientCertDER,
	}, nil
}
