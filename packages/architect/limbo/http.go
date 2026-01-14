// Package limbo...(TODO)
package limbo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/crypto"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/nix"
)

const shutdownTime time.Duration = 3 * time.Second

var done = make(chan error, 1)

func getTLSConfig(certs *crypto.Certificates) (*tls.Config, error) {
	caPool := x509.NewCertPool()
	caCert, err := x509.ParseCertificate(certs.CACertDER)
	if err != nil {
		return nil, err
	}
	caPool.AddCert(caCert)

	serverCert := tls.Certificate{
		Certificate: [][]byte{certs.CACertDER},
		PrivateKey:  certs.CAKeyPair.Priv,
	}

	return &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

func Descend(certs *crypto.Certificates, lport int, flake *nix.Flake) error {
	tlsConf, err := getTLSConfig(certs)
	if err != nil {
		return err
	}

	m := Manifest{flake: flake}

	r := mux.NewRouter()
	r.HandleFunc("/", m.handler).Methods("POST")
	r.HandleFunc("/flake", m.sendFlake).Methods("GET")
	r.HandleFunc("/closure", m.sendClosure).Methods("GET")
	r.HandleFunc("/diff", handleDiff).Methods("POST")
	r.HandleFunc("/nar/{hash}", handleNar).Methods("GET")
	r.HandleFunc("/status", handleStatus).Methods("POST")

	s := &http.Server{
		Addr:      ":" + strconv.Itoa(lport), /* FIXME: Use configured addr */
		Handler:   r,
		TLSConfig: tlsConf,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Error("server failed: %v", err)
			os.Exit(1)
		}
	}()
	log.Info("waiting for dreamer...")

	select {
	case <-c: /* NOTE: SIGINT/SIGTERM */
	case err := <-done: /* NOTE: Handler signaled completion */
		if err != nil {
			log.Error("dreamer failed: %v", err)
		} else {
			log.Highlight("dreamer finished installation")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTime)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("forced shutdown: %v", err)
	}

	return nil
}
