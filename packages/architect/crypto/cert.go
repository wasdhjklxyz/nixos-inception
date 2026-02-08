// Package crypto
package crypto

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"
)

type CAState struct {
	KeyPair *RSAKeyPair
	CertDER []byte
	Cert    *x509.Certificate
}

type ServerCert struct {
	KeyPair *RSAKeyPair
	CertDER []byte
}

type DreamerCert struct {
	KeyPair *RSAKeyPair
	CertDER []byte
}

type Certificates struct {
	CAKeyPair      *RSAKeyPair
	CACertDER      []byte
	DreamerKeyPair *RSAKeyPair
	DreamerCertDER []byte
}

const commonName = "nixos-inception"

func InitCA(stateDir string, dur, skew time.Duration) (*CAState, error) {
	return nil, nil
}

func InitServer(ca *CAState, stateDir string, dur, skew time.Duration) (*ServerCert, error) {
	return nil, nil
}

func IssueDreamerCert(ca *CAState, dur, skew time.Duration) (*DreamerCert, error) {
	return nil, nil
}

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
	}, nil
}

func createServerTemplate(dur, skew time.Duration) (*x509.Certificate, error) {
	now := time.Now().UTC()

	sn, err := generateSerialNumber()
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber: sn,
		Subject:      pkix.Name{CommonName: commonName + "-server"},

		NotBefore: now.Add(-skew),
		NotAfter:  now.Add(dur),

		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		IPAddresses: []net.IP{net.ParseIP("0.0.0.0")}, // FIXME: Use IP from deploy opts
	}, nil
}

func createDreamerTemplate(dur, skew time.Duration) (*x509.Certificate, error) {
	now := time.Now().UTC()

	sn, err := generateSerialNumber()
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber: sn,
		Subject:      pkix.Name{CommonName: commonName + "-dreamer"},

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

	dreamerKeyPair, err := GenerateRSAKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate dreamer RSA key pair: %v", err)
	}

	dreamerCertTemplate, err := createDreamerTemplate(dur, skew)
	if err != nil {
		return nil, fmt.Errorf("failed to create dreamer certificate template: %v", err)
	}

	dreamerCertDER, err := x509.CreateCertificate(
		rand.Reader,
		dreamerCertTemplate,
		caCert,
		dreamerKeyPair.pub,
		caKeyPair.Priv,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create dreamer certificate: %v", err)
	}

	return &Certificates{
		CAKeyPair:      caKeyPair,
		CACertDER:      caCertDER,
		DreamerKeyPair: dreamerKeyPair,
		DreamerCertDER: dreamerCertDER,
	}, nil
}
