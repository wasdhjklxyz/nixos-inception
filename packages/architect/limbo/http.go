// Package limbo...(TODO)
package limbo

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

const shutdownTime time.Duration = 3 * time.Second

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(os.Stderr, "GOT SOEMTHIGN!")
	w.Write([]byte("hello world"))
}

func StartHTTPListener(lport int) error {
	r := mux.NewRouter()
	r.HandleFunc("/", helloWorld).Methods("GET")

	s := &http.Server{
		Addr:    ":" + strconv.Itoa(lport),
		Handler: r,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server failed: %v", err)
			os.Exit(1)
		}
	}()

	<-c // Blocks until signal received
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTime)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server forced to shutdown: %v", err)
	}

	return nil
}
