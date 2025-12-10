package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	/* TODO: Strings also in lib/default.nix should use one source */
	certPath   = "/etc/nixos-inception/dreamer.crt"
	keyPath    = "/etc/nixos-inception/dreamer.key"
	caPath     = "/etc/nixos-inception/ca.crt"
	configPath = "/etc/nixos-inception/config"
)

func LoadIdentity() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	caBytes, err := os.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caBytes) {
		return nil, err
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caPool,
		InsecureSkipVerify: true, /* FIXME */
		MinVersion:         tls.VersionTLS13,
	}, nil
}

func Reach(addr string) error {
	tlsConf, err := LoadIdentity()
	if err != nil {
		return err
	}

	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConf},
		Timeout:   30 * time.Second,
	}

	for {
		resp, err := client.Get(addr)
		if err != nil {
			time.Sleep(5 * time.Second) /* TODO: Make configurable */
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		fmt.Println(string(body))
		return nil
	}
}

func main() {
	addr, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	url := fmt.Sprintf("https://%s", strings.TrimSpace(string(addr)))

	if err := Reach(url); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
