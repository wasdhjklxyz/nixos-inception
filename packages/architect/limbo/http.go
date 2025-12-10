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
)

const shutdownTime time.Duration = 3 * time.Second

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(os.Stderr, "GOT SOEMTHIGN!")
	w.Write([]byte("hello world"))
}

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

func Descend(certs *crypto.Certificates, lport int) error {
	tlsConf, err := getTLSConfig(certs)
	if err != nil {
		return err
	}

	r := mux.NewRouter()
	r.HandleFunc("/", helloWorld).Methods("GET")

	s := &http.Server{
		Addr:      ":" + strconv.Itoa(lport), /* FIXME: Use configured addr */
		Handler:   r,
		TLSConfig: tlsConf,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server failed: %v", err)
			os.Exit(1)
		}
	}()

	<-c // Blocks until signal received
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTime)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server forced to shutdown: %v", err)
		return err
	}

	return nil
}
