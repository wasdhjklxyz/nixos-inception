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
)

const shutdownTime time.Duration = 3 * time.Second

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

func Descend(certs *crypto.Certificates, lport int, closure *Closure) error {
	tlsConf, err := getTLSConfig(certs)
	if err != nil {
		return err
	}

	r := mux.NewRouter()
	r.HandleFunc("/", closure.handler).Methods("POST")
	r.HandleFunc("/diff", handleDiff).Methods("POST")
	r.HandleFunc("/nar/{hash}", handleNar).Methods("GET")
	r.HandleFunc("/nar-done", handleNarDone).Methods("POST")

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

	<-c // Blocks until signal received
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTime)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("forced shutdown: %v", err)
	}

	return nil
}
