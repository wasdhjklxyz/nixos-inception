// Package limbo...(TODO)
package limbo

import (
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
)

var (
	totalPaths  int
	currentPath int
	totalBytes  int64
	sentBytes   int64
)

type countingWriter struct {
	w http.ResponseWriter
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	atomic.AddInt64(&sentBytes, int64(n))
	return n, err
}

func handleNar(w http.ResponseWriter, r *http.Request) {
	hash := mux.Vars(r)["hash"]
	path := "/nix/store/" + hash

	if _, err := os.Stat(path); err != nil {
		log.Warn("nar request for missing path: %s", hash)
		http.Error(w, "not found", 404)
		return
	}

	/* FIXME: Dreamer and architect to use same constants for headers */
	if tb := r.Header.Get("Inception-Total-Bytes"); tb != "" {
		totalBytes, _ = strconv.ParseInt(tb, 10, 64) /* TODO: Error handling */
	}
	if t := r.Header.Get("Inception-Total"); t != "" {
		totalPaths, _ = strconv.Atoi(t) /* TODO: Error handling */
	}
	if c := r.Header.Get("Inception-Current"); c != "" {
		currentPath, _ = strconv.Atoi(c) /* TODO: Error handling */
	}

	/* Perhaps a bit extra but at least it looks nice and looks like nix output */
	done := make(chan bool)
	go func() {
		log.Progress(
			log.ProgressState{
				Done:       currentPath - 1,
				Running:    1,
				Total:      totalPaths,
				Bytes:      atomic.LoadInt64(&sentBytes),
				TotalBytes: totalBytes,
			}, path,
		)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				log.Progress(
					log.ProgressState{
						Done:       currentPath - 1,
						Running:    1,
						Total:      totalPaths,
						Bytes:      atomic.LoadInt64(&sentBytes),
						TotalBytes: totalBytes,
					}, path,
				)
			}
		}
	}()

	counter := &countingWriter{w: w}
	var stderr bytes.Buffer
	cmd := exec.Command("nix-store", "--export", path)
	cmd.Stdout = counter
	cmd.Stderr = &stderr
	err := cmd.Run()

	done <- true

	if err != nil {
		log.ProgressDone()
		log.Error("export %s: %v: %s", hash, err, stderr.String())
	}
}

func handleNarDone(w http.ResponseWriter, _ *http.Request) {
	log.ProgressDone()
	w.WriteHeader(http.StatusNoContent)
}
