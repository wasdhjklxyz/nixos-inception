package main

import "github.com/gorilla/mux"

type server struct {
	config *config
	router *mux.Router
}

func (s *server) run() {
}

func newServer(cfg *config) (*server, error) {
	r := mux.NewRouter()
	return &server{config: cfg, router: r}, nil
}
