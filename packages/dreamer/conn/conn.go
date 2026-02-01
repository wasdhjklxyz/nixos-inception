// Package conn...(TODO)
package conn

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	addrPath = "/etc/nixos-inception/config"
	certPath = "/etc/nixos-inception/dreamer.crt"
	keyPath  = "/etc/nixos-inception/dreamer.key"
	caPath   = "/etc/nixos-inception/ca.crt"
	timeout  = 30 * time.Second
)

type Conn struct {
	url    string
	client *http.Client
}

func NewConn() (c *Conn, err error) {
	c = &Conn{}
	if c.url, err = getURL(); err != nil {
		return
	}
	if c.client, err = newHTTPClient(); err != nil {
		return
	}
	if err = c.wait(); err != nil {
		return
	}
	return
}

func getURL() (string, error) {
	addr, err := os.ReadFile(addrPath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://%s", strings.TrimSpace(string(addr))), nil
}

func newHTTPClient() (*http.Client, error) {
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
		Timeout: 0, /* NOTE: No timeout FIXME: Use timeout const */
	}, nil
}

func (c *Conn) wait() error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := c.GetHealth(); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("server not reachable after %v", timeout)
}
