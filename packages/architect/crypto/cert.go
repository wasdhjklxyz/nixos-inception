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

const commonName = "nixos-inception"

func InitCA(stateDir string, dur, skew time.Duration) (*CAState, error) {
	keyPair, err := GenerateRSAKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate CA RSA key pair: %v", err)
	}

	certTmpl, err := createCATemplate(dur, skew)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate template: %v", err)
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader,
		certTmpl,
		certTmpl,
		keyPair.pub,
		keyPair.Priv,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate DER: %v", err)
	}

	return &CAState{keyPair, certDER, cert}, nil
}

func InitServer(ca *CAState, stateDir string, dur, skew time.Duration) (*ServerCert, error) {
	return nil, nil
}

func IssueDreamerCert(ca *CAState, dur, skew time.Duration) (*DreamerCert, error) {
	keyPair, err := GenerateRSAKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate dreamer RSA key pair: %v", err)
	}

	certTmpl, err := createDreamerTemplate(dur, skew)
	if err != nil {
		return nil, fmt.Errorf("failed to create dreamer certificate template: %v", err)
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader,
		certTmpl,
		ca.Cert,
		keyPair.pub,
		ca.KeyPair.Priv,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create dreamer certificate: %v", err)
	}

	return &DreamerCert{keyPair, certDER}, nil
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

		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}, nil
}

func createServerTemplate(serverIP net.IP, dur, skew time.Duration) (*x509.Certificate, error) {
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

		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		IPAddresses: []net.IP{serverIP},
	}, nil
}

func createDreamerTemplate(dur, skew time.Duration) (*x509.Certificate, error) {
	now := time.Now().UTC()

	sn, err := generateSerialNumber()
	if err != nil {
		return nil, err
	}

	// TODO: Common name should include dreamer ID for logs, revocation, etc.
	return &x509.Certificate{
		SerialNumber: sn,
		Subject:      pkix.Name{CommonName: commonName + "-dreamer"},

		NotBefore: now.Add(-skew),
		NotAfter:  now.Add(dur),

		KeyUsage: x509.KeyUsageDigitalSignature,

		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}, nil
}
