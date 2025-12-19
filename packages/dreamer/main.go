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

func newClient() (*http.Client, error) {
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

	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caPool,
			InsecureSkipVerify: true, /* FIXME */
			MinVersion:         tls.VersionTLS13,
		}},
		Timeout: 30 * time.Second, /* TODO: Make configurable */
	}, nil
}

func reach(client *http.Client, url string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println(string(body))
	return nil
}

func main() {
	addr, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	client, err := newClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	url := fmt.Sprintf("https://%s", strings.TrimSpace(string(addr)))

	if err := reach(client, url); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
