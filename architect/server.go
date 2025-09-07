package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type server struct {
	http.Server
	config *config
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hi"))
}

func (s *server) run() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server failed: %v", err)
			os.Exit(1)
		}
	}()

	<-c // Block until signal received
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server forced to shutdown: %v", err)
	}
}

func newServer(cfg *config) (*server, error) {
	r := mux.NewRouter()
	r.HandleFunc("/", authHandler)

	return &server{
		Server: http.Server{
			Addr:    cfg.addr,
			Handler: r,
		},
		config: cfg,
	}, nil
}
