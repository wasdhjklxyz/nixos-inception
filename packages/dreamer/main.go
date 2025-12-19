package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
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

type Closure struct {
	TopLevel   string   `json:"toplevel"`
	Requisites []string `json:"requisites"`
}

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

func fetchClosure(client *http.Client, url string) (*Closure, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var c Closure
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func diffClosure(requisites []string) ([]string, error) {
	ents, err := os.ReadDir("/nix/store")
	if err != nil {
		return nil, err
	}

	have := make(map[string]bool)
	for _, e := range ents {
		have[e.Name()] = true
	}

	var need []string
	for _, path := range requisites {
		name := strings.TrimPrefix(path, "/nix/store")
		if !have[name] {
			need = append(need, path)
		}
	}
	return need, nil
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

	c, err := fetchClosure(client, url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	need, err := diffClosure(c.Requisites)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(need)
}
