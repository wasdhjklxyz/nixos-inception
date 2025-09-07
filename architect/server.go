package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type server struct {
	config *config
	router *mux.Router
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hi"))
}

func (s *server) run() {
	http.ListenAndServe(s.config.addr, s.router)
}

func newServer(cfg *config) (*server, error) {
	r := mux.NewRouter()
	r.HandleFunc("/", authHandler)
	return &server{config: cfg, router: r}, nil
}
